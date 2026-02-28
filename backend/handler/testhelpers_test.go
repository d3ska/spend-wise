package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/auth"
	"backend/model"

	"github.com/go-chi/chi/v5"
)

// newAuthenticatedRequest creates an httptest.Request with a user ID in the auth context.
func newAuthenticatedRequest(t testing.TB, method, path string, body io.Reader, userID model.UserID) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, path, body)
	ctx := auth.WithUserID(r.Context(), userID)
	return r.WithContext(ctx)
}

// withChiParam adds a chi URL parameter to the request context.
func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// decodeResponseBody reads and JSON-decodes the response body into dst.
func decodeResponseBody(t testing.TB, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dst); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
}
