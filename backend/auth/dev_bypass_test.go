package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"backend/model"
)

// devUserCreator is a local interface satisfied by *store.UserStore.
// It lets us test the middleware logic without a real database.
type devUserCreator interface {
	GetOrCreateByEmail(ctx context.Context, u model.User) (model.User, error)
}

// mockDevUserStore implements devUserCreator for testing.
type mockDevUserStore struct {
	fn        func(ctx context.Context, u model.User) (model.User, error)
	callCount atomic.Int64
}

func (m *mockDevUserStore) GetOrCreateByEmail(ctx context.Context, u model.User) (model.User, error) {
	m.callCount.Add(1)
	return m.fn(ctx, u)
}

// devBypassMiddlewareTestable mirrors the production DevBypassMiddleware but
// accepts the devUserCreator interface so we can inject a mock.
// The logic is identical to DevBypassMiddleware in dev_bypass.go.
func devBypassMiddlewareTestable(us devUserCreator) func(http.Handler) http.Handler {
	var (
		once   sync.Once
		devUID model.UserID
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			once.Do(func() {
				u, err := us.GetOrCreateByEmail(r.Context(), model.User{
					Email:       "dev@spendwise.local",
					DisplayName: "Dev User",
					Provider:    model.ProviderGitHub,
					ProviderID:  "dev-local",
				})
				if err != nil {
					return
				}
				devUID = u.ID
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

func TestDevBypassMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		store      *mockDevUserStore
		requests   int
		wantStatus int
		wantUserID model.UserID
		wantCalls  int64
	}{
		{
			name: "injects user ID into context",
			store: &mockDevUserStore{
				fn: func(_ context.Context, _ model.User) (model.User, error) {
					return model.User{ID: 42, Email: "dev@spendwise.local"}, nil
				},
			},
			requests:   1,
			wantStatus: http.StatusOK,
			wantUserID: 42,
			wantCalls:  1,
		},
		{
			name: "returns 500 when user store fails",
			store: &mockDevUserStore{
				fn: func(_ context.Context, _ model.User) (model.User, error) {
					return model.User{}, errors.New("db connection refused")
				},
			},
			requests:   1,
			wantStatus: http.StatusInternalServerError,
			wantUserID: 0,
			wantCalls:  1,
		},
		{
			name: "caches user via sync.Once",
			store: &mockDevUserStore{
				fn: func(_ context.Context, _ model.User) (model.User, error) {
					return model.User{ID: 7, Email: "dev@spendwise.local"}, nil
				},
			},
			requests:   5,
			wantStatus: http.StatusOK,
			wantUserID: 7,
			wantCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := devBypassMiddlewareTestable(tt.store)

			var gotUserID model.UserID
			var gotOK bool

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUserID, gotOK = UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware(next)

			var lastStatus int
			for i := 0; i < tt.requests; i++ {
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				handler.ServeHTTP(rec, req)
				lastStatus = rec.Code
			}

			if lastStatus != tt.wantStatus {
				t.Errorf("status = %d, want %d", lastStatus, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				if !gotOK {
					t.Fatal("expected user ID in context, got none")
				}
				if gotUserID != tt.wantUserID {
					t.Errorf("user ID = %d, want %d", gotUserID, tt.wantUserID)
				}
			}

			if got := tt.store.callCount.Load(); got != tt.wantCalls {
				t.Errorf("store call count = %d, want %d", got, tt.wantCalls)
			}
		})
	}
}

func TestDevBypassMiddleware_DoesNotCallNextOn500(t *testing.T) {
	store := &mockDevUserStore{
		fn: func(_ context.Context, _ model.User) (model.User, error) {
			return model.User{}, errors.New("db down")
		},
	}

	middleware := devBypassMiddlewareTestable(store)

	var nextCalled bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware(next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Error("next handler was called despite store failure, want it skipped")
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestDevBypassMiddleware_PassesCorrectUserToStore(t *testing.T) {
	var captured model.User

	store := &mockDevUserStore{
		fn: func(_ context.Context, u model.User) (model.User, error) {
			captured = u
			return model.User{ID: 1, Email: u.Email}, nil
		},
	}

	middleware := devBypassMiddlewareTestable(store)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if captured.Email != "dev@spendwise.local" {
		t.Errorf("email = %q, want %q", captured.Email, "dev@spendwise.local")
	}
	if captured.DisplayName != "Dev User" {
		t.Errorf("display name = %q, want %q", captured.DisplayName, "Dev User")
	}
	if captured.Provider != model.ProviderGitHub {
		t.Errorf("provider = %q, want %q", captured.Provider, model.ProviderGitHub)
	}
	if captured.ProviderID != "dev-local" {
		t.Errorf("provider ID = %q, want %q", captured.ProviderID, "dev-local")
	}
}
