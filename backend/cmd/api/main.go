package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/auth"
	"backend/banksync"
	"backend/config"

	"backend/handler"
	"backend/model"
	"backend/service"
	"backend/store"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Load .env file if present (env vars take precedence).
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx := context.Background()

	pool, err := store.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	// OAuth providers.
	providers := map[model.AuthProvider]auth.Provider{
		model.ProviderGoogle: auth.NewGoogleProvider(
			cfg.Auth.GoogleClientID,
			cfg.Auth.GoogleClientSecret,
			cfg.Auth.GoogleRedirectURL,
		),
		model.ProviderGitHub: auth.NewGitHubProvider(
			cfg.Auth.GitHubClientID,
			cfg.Auth.GitHubClientSecret,
			cfg.Auth.GitHubRedirectURL,
		),
	}

	// Stores.
	userStore := store.NewUserStore(pool)
	workspaceStore := store.NewWorkspaceStore(pool)
	inviteStore := store.NewInviteStore(pool)
	categoryStore := store.NewCategoryStore(pool)
	transactionStore := store.NewTransactionStore(pool)
	summaryStore := store.NewSummaryStore(pool)
	fundingStore := store.NewFundingStore(pool)
	ruleStore := store.NewRuleStore(pool)
	bankConnectionStore := store.NewBankConnectionStore(pool)
	bankAccountStore := store.NewBankAccountStore(pool)

	// Enable Banking client.
	enableBankingClient, err := banksync.NewHTTPClient(cfg.EnableBanking.BaseURL, cfg.EnableBanking.ApplicationID, cfg.EnableBanking.KeyPath)
	if err != nil {
		return fmt.Errorf("creating Enable Banking client: %w", err)
	}

	// Services.
	authService := service.NewAuthService(providers, userStore, cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)
	workspaceService := service.NewWorkspaceService(workspaceStore, categoryStore)
	inviteService := service.NewInviteService(inviteStore, workspaceStore, cfg.Invite.DefaultExpiryHours)
	ruleService := service.NewRuleService(ruleStore, workspaceStore)
	transactionService := service.NewTransactionService(transactionStore, workspaceStore, categoryStore, ruleService)
	summaryService := service.NewSummaryService(summaryStore, fundingStore, workspaceStore)
	fundingService := service.NewFundingService(fundingStore, workspaceStore)
	bankConnectionService := service.NewBankConnectionService(bankConnectionStore, bankAccountStore, enableBankingClient)
	bankAccountLinkingService := service.NewBankAccountLinkingService(bankAccountStore, bankConnectionStore, workspaceStore, transactionStore)
	bankSyncService := service.NewBankSyncService(bankConnectionStore, bankAccountStore, transactionStore, categoryStore, ruleStore, enableBankingClient)

	// SSE hub.
	sseHub := handler.NewSSEHub()

	// Handlers.
	cookieCfg := handler.CookieConfig{
		Domain: cfg.Auth.CookieDomain,
		Secure: cfg.Auth.CookieSecure,
		MaxAge: int(cfg.Auth.JWTExpiry.Seconds()),
	}
	oauthStateStore := auth.NewStateStore(10 * time.Minute)
	authHandler := handler.NewAuthHandler(authService, providers, userStore, cookieCfg, oauthStateStore)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService, sseHub)
	inviteHandler := handler.NewInviteHandler(inviteService, workspaceService, sseHub)
	categoryHandler := handler.NewCategoryHandler(workspaceService, sseHub)
	transactionHandler := handler.NewTransactionHandler(transactionService, sseHub)
	summaryHandler := handler.NewSummaryHandler(summaryService)
	fundingHandler := handler.NewFundingHandler(fundingService, sseHub)
	ruleHandler := handler.NewRuleHandler(ruleService, sseHub)
	bankHandler := handler.NewBankHandler(bankConnectionService, bankAccountLinkingService, bankSyncService)
	sseHandler := handler.NewSSEHandler(sseHub, workspaceService)

	// Rate limiter for auth endpoints: 10 requests/second per IP, burst of 20.
	authRateLimiter := auth.NewRateLimiter(10, 20)

	routerCfg := handler.RouterConfig{
		AuthHandler:        authHandler,
		WorkspaceHandler:   workspaceHandler,
		InviteHandler:      inviteHandler,
		CategoryHandler:    categoryHandler,
		TransactionHandler: transactionHandler,
		SummaryHandler:     summaryHandler,
		FundingHandler:     fundingHandler,
		RuleHandler:        ruleHandler,
		BankHandler:        bankHandler,
		SSEHub:             sseHub,
		SSEHandler:         sseHandler,
		JWTSecret:          cfg.Auth.JWTSecret,
		AuthBypass:         cfg.Auth.AuthBypass,
		CORSAllowedOrigins: cfg.CORS.AllowedOrigins,
		AuthRateLimiter:    auth.RateLimitMiddleware(authRateLimiter),
	}

	if cfg.Auth.AuthBypass {
		slog.Warn("AUTH_BYPASS is enabled — all requests will use a dev user, do NOT use in production")
		routerCfg.DevBypassMW = auth.DevBypassMiddleware(userStore)
	}

	router := handler.NewRouter(routerCfg)

	// Cron scheduler for bank sync.
	cronScheduler := cron.New()
	_, err = cronScheduler.AddFunc("0 */6 * * *", func() {
		slog.Info("running scheduled bank sync")
		bankSyncService.SyncAll(context.Background())
	})
	if err != nil {
		return fmt.Errorf("registering bank sync cron job: %w", err)
	}
	cronScheduler.Start()
	slog.Info("cron scheduler started", "schedule", "every 6 hours")

	srv := &http.Server{
		Addr:              cfg.Server.Addr(),
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       120 * time.Second,
	}

	// Start server in a goroutine.
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for shutdown signal or server error.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	// Graceful shutdown: stop cron, close SSE hub, then HTTP server.
	slog.Info("stopping cron scheduler")
	cronScheduler.Stop()

	slog.Info("closing SSE hub")
	sseHub.Close()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	slog.Info("server stopped gracefully")
	return nil
}
