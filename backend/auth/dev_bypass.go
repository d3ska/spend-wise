package auth

import (
	"log/slog"
	"net/http"
	"sync"

	"backend/model"
	"backend/store"
)

// DevBypassMiddleware returns Chi middleware that skips JWT validation and
// injects a hardcoded dev user into the request context. The dev user is
// created (or fetched) on the first request and cached for subsequent ones.
//
// This must only be used for local development.
func DevBypassMiddleware(userStore *store.UserStore) func(http.Handler) http.Handler {
	var (
		once   sync.Once
		devUID model.UserID
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			once.Do(func() {
				u, err := userStore.GetOrCreateByEmail(r.Context(), model.User{
					Email:       "dev@spendwise.local",
					DisplayName: "Dev User",
					Provider:    model.ProviderGitHub,
					ProviderID:  "dev-local",
				})
				if err != nil {
					slog.Error("dev bypass: failed to create dev user", "error", err)
					return
				}
				devUID = u.ID
				slog.Info("dev bypass: using dev user", "id", devUID, "email", u.Email)
			})

			if devUID == 0 {
				http.Error(w, `{"error":"dev bypass: failed to create dev user"}`, http.StatusInternalServerError)
				return
			}

			ctx := WithUserID(r.Context(), devUID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
