package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/auth"
)

func TestAuthHandler_HandleLogout(t *testing.T) {
	h := &AuthHandler{
		cookie: CookieConfig{Domain: "localhost", Secure: false, MaxAge: 3600},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()

	h.HandleLogout(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
	}

	cookies := rec.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == auth.CookieName {
			found = true
			if c.MaxAge != -1 {
				t.Errorf("got MaxAge %d, want -1", c.MaxAge)
			}
			if c.Value != "" {
				t.Errorf("got Value %q, want empty", c.Value)
			}
		}
	}
	if !found {
		t.Error("expected cookie to be set, got none")
	}
}
