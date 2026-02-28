package handler

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"

	"backend/auth"
	"backend/model"
	"backend/service"
	"backend/store"

	"github.com/go-chi/chi/v5"
)

// CookieConfig holds settings for the auth session cookie.
type CookieConfig struct {
	Domain string
	Secure bool
	MaxAge int // seconds
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
	providers   map[model.AuthProvider]auth.Provider
	userStore   *store.UserStore
	cookie      CookieConfig
	stateStore  *auth.StateStore
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(
	authService *service.AuthService,
	providers map[model.AuthProvider]auth.Provider,
	userStore *store.UserStore,
	cookie CookieConfig,
	stateStore *auth.StateStore,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		providers:   providers,
		userStore:   userStore,
		cookie:      cookie,
		stateStore:  stateStore,
	}
}

// GetAuthURL returns the OAuth2 authorization URL for the given provider.
func (h *AuthHandler) GetAuthURL(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.providers[model.AuthProvider(providerName)]
	if !ok {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "unsupported provider: "+providerName)
		return
	}

	state, err := generateState()
	if err != nil {
		respondCodedError(w, http.StatusInternalServerError, model.CodeInternalError, "failed to generate state")
		return
	}

	h.stateStore.Store(state)

	url := provider.AuthCodeURL(state)
	respondJSON(w, http.StatusOK, map[string]string{
		"url":   url,
		"state": state,
	})
}

// HandleCallback exchanges an OAuth code for a JWT and user profile.
// The JWT is set as an httpOnly cookie; the response body contains only the user.
func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")

	var body struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}
	if body.Code == "" {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "code is required")
		return
	}
	if body.State == "" || !h.stateStore.Validate(body.State) {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid or expired state parameter")
		return
	}

	token, user, err := h.authService.Authenticate(r.Context(), providerName, body.Code)
	if err != nil {
		slog.Error("authentication failed", "provider", providerName, "error", err)
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "authentication failed")
		return
	}

	http.SetCookie(w, h.newAuthCookie(token))

	respondJSON(w, http.StatusOK, map[string]any{
		"user": userResponse(user),
	})
}

// HandleLogout clears the auth cookie.
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, _ *http.Request) {
	c := h.newAuthCookie("")
	c.MaxAge = -1
	http.SetCookie(w, c)

	respondJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *AuthHandler) newAuthCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     auth.CookieName,
		Value:    value,
		Path:     "/api",
		Domain:   h.cookie.Domain,
		MaxAge:   h.cookie.MaxAge,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// GetMe returns the currently authenticated user.
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	user, err := h.userStore.GetByID(r.Context(), userID)
	if err != nil {
		respondCodedError(w, http.StatusNotFound, model.ErrUserNotFound.Code, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, userResponse(user))
}

// UpdateProfile handles PATCH /api/v1/auth/me.
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	var body struct {
		PreferredLanguage *string `json:"preferred_language"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	// If no fields provided, return current profile (no-op).
	if body.PreferredLanguage == nil {
		user, err := h.userStore.GetByID(r.Context(), userID)
		if err != nil {
			respondCodedError(w, http.StatusNotFound, model.ErrUserNotFound.Code, "user not found")
			return
		}
		respondJSON(w, http.StatusOK, userResponse(user))
		return
	}

	user, err := h.authService.UpdatePreferredLanguage(r.Context(), userID, *body.PreferredLanguage)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, userResponse(user))
}

func userResponse(u model.User) map[string]any {
	return map[string]any{
		"id":                 u.ID,
		"email":              u.Email,
		"display_name":       u.DisplayName,
		"avatar_url":         u.AvatarURL,
		"preferred_language": u.PreferredLanguage,
	}
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
