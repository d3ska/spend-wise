package auth

import (
	"net/http"
	"strings"
)

const CookieName = "sw_token"

// JWTMiddleware returns Chi middleware that validates a JWT token
// and injects the UserID into the request context.
// It first checks for the sw_token cookie, then falls back to
// the Authorization: Bearer header.
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := tokenFromRequest(r)
			if tokenStr == "" {
				http.Error(w, `{"error":"missing or invalid token"}`, http.StatusUnauthorized)
				return
			}

			userID, err := VerifyToken(tokenStr, secret)
			if err != nil {
				http.Error(w, `{"error":"missing or invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// tokenFromRequest extracts the JWT token from the request, checking
// the sw_token cookie first, then the Authorization: Bearer header.
func tokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(CookieName); err == nil && c.Value != "" {
		return c.Value
	}
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}
