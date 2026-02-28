package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/model"
)

func TestJWTMiddleware(t *testing.T) {
	const secret = "test-secret"
	const userID model.UserID = 42

	validToken, err := SignToken(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("signing token: %v", err)
	}

	handler := JWTMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected user ID in context")
		}
		if id != userID {
			t.Errorf("expected user ID %d, got %d", userID, id)
		}
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name       string
		setup      func(r *http.Request)
		wantStatus int
	}{
		{
			name: "cookie only",
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: CookieName, Value: validToken})
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "bearer only",
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer "+validToken)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "both present — cookie wins",
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: CookieName, Value: validToken})
				r.Header.Set("Authorization", "Bearer invalid-token")
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "neither present",
			setup:      func(r *http.Request) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid cookie token",
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: CookieName, Value: "invalid"})
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid bearer token",
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer invalid")
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			tt.setup(req)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
