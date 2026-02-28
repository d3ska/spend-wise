package handler

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"backend/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// RouterConfig holds dependencies needed to build the router.
type RouterConfig struct {
	AuthHandler        *AuthHandler
	WorkspaceHandler   *WorkspaceHandler
	InviteHandler      *InviteHandler
	CategoryHandler    *CategoryHandler
	TransactionHandler *TransactionHandler
	SummaryHandler     *SummaryHandler
	FundingHandler     *FundingHandler
	RuleHandler        *RuleHandler
	BankHandler        *BankHandler
	SSEHub             *SSEHub
	SSEHandler         *SSEHandler
	JWTSecret          string
	AuthBypass         bool
	DevBypassMW        func(http.Handler) http.Handler
	CORSAllowedOrigins string
	AuthRateLimiter    func(http.Handler) http.Handler
}

// NewRouter creates a Chi router with global middleware, healthz, and auth routes.
func NewRouter(cfg RouterConfig) chi.Router {
	r := chi.NewRouter()

	r.Use(securityHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   parseOrigins(cfg.CORSAllowedOrigins),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(safeRecoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", handleHealthz)

	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes (rate-limited to prevent brute force).
		r.Route("/auth/{provider}", func(r chi.Router) {
			if cfg.AuthRateLimiter != nil {
				r.Use(cfg.AuthRateLimiter)
			}
			r.Get("/url", cfg.AuthHandler.GetAuthURL)
			r.Post("/callback", cfg.AuthHandler.HandleCallback)
		})
		r.Post("/auth/logout", cfg.AuthHandler.HandleLogout)

		// Public invite preview.
		r.Get("/invites/{code}", cfg.InviteHandler.PreviewInvite)

		// Protected routes.
		r.Group(func(r chi.Router) {
			if cfg.AuthBypass && cfg.DevBypassMW != nil {
				r.Use(cfg.DevBypassMW)
			} else {
				r.Use(auth.JWTMiddleware(cfg.JWTSecret))
			}
			r.Get("/auth/me", cfg.AuthHandler.GetMe)
			r.Patch("/auth/me", cfg.AuthHandler.UpdateProfile)

			// Accept invite (requires auth).
			r.Post("/invites/{code}/accept", cfg.InviteHandler.AcceptInvite)

			// Bank institution routes.
			r.Get("/banks", cfg.BankHandler.ListInstitutions)

			// All user bank accounts (across all connections).
			r.Get("/bank-accounts", cfg.BankHandler.ListUserBankAccounts)
			r.Patch("/bank-accounts/{id}", cfg.BankHandler.UpdateBankAccountCustomName)

			// Bank connection routes.
			r.Post("/bank-connections", cfg.BankHandler.InitiateConnection)
			r.Get("/bank-connections", cfg.BankHandler.ListConnections)
			r.Post("/bank-connections/complete", cfg.BankHandler.CompleteConnectionByCode)
			r.Route("/bank-connections/{id}", func(r chi.Router) {
				r.Post("/complete", cfg.BankHandler.CompleteConnection)
				r.Delete("/", cfg.BankHandler.DeleteConnection)
				r.Post("/deactivate", cfg.BankHandler.DeactivateConnection)
				r.Post("/reactivate", cfg.BankHandler.ReactivateConnection)
				r.Post("/reconnect", cfg.BankHandler.ReconnectConnection)
				r.Post("/sync", cfg.BankHandler.TriggerSync)
			})

			// Workspace routes.
			r.Post("/workspaces", cfg.WorkspaceHandler.CreateWorkspace)
			r.Get("/workspaces", cfg.WorkspaceHandler.ListWorkspaces)
			r.Route("/workspaces/{id}", func(r chi.Router) {
				r.Get("/events", cfg.SSEHandler.StreamEvents)
				r.Get("/", cfg.WorkspaceHandler.GetWorkspace)
				r.Put("/", cfg.WorkspaceHandler.UpdateWorkspace)
				r.Delete("/", cfg.WorkspaceHandler.DeleteWorkspace)

				// Member routes.
				r.Get("/members", cfg.InviteHandler.ListMembers)
				r.Post("/members", cfg.WorkspaceHandler.AddMember)
				r.Put("/members/{userID}", cfg.InviteHandler.UpdateMemberRole)
				r.Delete("/members/{userID}", cfg.WorkspaceHandler.RemoveMember)

				// Invite routes.
				r.Post("/invites", cfg.InviteHandler.CreateInvite)

				// Category routes.
				r.Post("/categories", cfg.CategoryHandler.CreateCategory)
				r.Get("/categories", cfg.CategoryHandler.ListCategories)
				r.Put("/categories/{catID}", cfg.CategoryHandler.UpdateCategory)
				r.Delete("/categories/{catID}", cfg.CategoryHandler.DeleteCategory)

				// Transaction routes.
				r.Post("/transactions", cfg.TransactionHandler.CreateTransaction)
				r.Get("/transactions", cfg.TransactionHandler.ListTransactions)
				r.Post("/transactions/import", cfg.TransactionHandler.ImportTransactions)
				r.Post("/transactions/bulk-categorize", cfg.TransactionHandler.BulkCategorizeTransactions)
				r.Post("/transactions/bulk-delete", cfg.TransactionHandler.BulkDeleteTransactions)
				r.Post("/transactions/apply-rules", cfg.TransactionHandler.ApplyRules)
				r.Get("/transactions/{txID}", cfg.TransactionHandler.GetTransaction)
				r.Put("/transactions/{txID}", cfg.TransactionHandler.UpdateTransaction)
				r.Delete("/transactions/{txID}", cfg.TransactionHandler.DeleteTransaction)

				// Summary route.
				r.Get("/summary", cfg.SummaryHandler.GetSummary)

				// Funding routes.
				r.Put("/fundings", cfg.FundingHandler.RecordFunding)
				r.Get("/fundings", cfg.FundingHandler.ListFundings)
				r.Delete("/fundings/{fundingId}", cfg.FundingHandler.DeleteFunding)

				// Rule routes.
				r.Post("/rules", cfg.RuleHandler.CreateRule)
				r.Get("/rules", cfg.RuleHandler.ListRules)
				r.Put("/rules/{ruleID}", cfg.RuleHandler.UpdateRule)
				r.Delete("/rules/{ruleID}", cfg.RuleHandler.DeleteRule)
				r.Patch("/rules/{ruleID}/toggle", cfg.RuleHandler.ToggleRule)

				// Bank account linking routes.
				r.Post("/bank-accounts", cfg.BankHandler.LinkBankAccount)
				r.Get("/bank-accounts", cfg.BankHandler.ListLinkedBankAccounts)
				r.Delete("/bank-accounts/{bankAccountId}", cfg.BankHandler.UnlinkBankAccount)
			})
		})
	})

	return r
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// securityHeaders sets common security headers on every response.
// These mitigate clickjacking, MIME-sniffing, and restrict browser features.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

// safeRecoverer recovers from panics and returns a generic 500 response.
// Unlike chi's default Recoverer, it never leaks stack traces or panic values
// to the client. The full stack is logged server-side for debugging.
func safeRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					panic(rvr) // preserve ErrAbortHandler semantics
				}
				reqID := middleware.GetReqID(r.Context())
				slog.Error("panic recovered",
					"request_id", reqID,
					"method", r.Method,
					"path", r.URL.Path,
					"panic", rvr,
					"stack", string(debug.Stack()),
				)
				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// parseOrigins splits a comma-separated origins string into a slice.
// Wildcard ("*") origins are rejected because AllowCredentials is enabled,
// which makes a wildcard both insecure and invalid per the CORS spec.
func parseOrigins(s string) []string {
	var origins []string
	for _, o := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			slog.Warn("CORS: wildcard origin '*' rejected — use explicit origins with AllowCredentials")
			continue
		}
		origins = append(origins, trimmed)
	}
	return origins
}
