package service

import (
	"context"
	"fmt"
	"time"

	"backend/auth"
	"backend/model"
)

// SupportedLanguages is the set of allowed preferred_language values.
var SupportedLanguages = map[string]bool{
	"en": true,
	"pl": true,
}

// UserStoreIface defines the user store methods used by AuthService.
type UserStoreIface interface {
	GetOrCreateByEmail(ctx context.Context, u model.User) (model.User, error)
	GetByID(ctx context.Context, id model.UserID) (model.User, error)
	UpdatePreferredLanguage(ctx context.Context, id model.UserID, lang string) (model.User, error)
}

// AuthService handles OAuth2 authentication and JWT issuance.
type AuthService struct {
	providers map[model.AuthProvider]auth.Provider
	userStore UserStoreIface
	jwtSecret string
	jwtExpiry time.Duration
}

// NewAuthService creates an AuthService.
func NewAuthService(
	providers map[model.AuthProvider]auth.Provider,
	userStore UserStoreIface,
	jwtSecret string,
	jwtExpiry time.Duration,
) *AuthService {
	return &AuthService{
		providers: providers,
		userStore: userStore,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Authenticate exchanges an OAuth code for a user profile, creates or finds the
// user by email, and returns a signed JWT along with the user.
func (s *AuthService) Authenticate(ctx context.Context, providerName string, code string) (string, model.User, error) {
	provider, ok := s.providers[model.AuthProvider(providerName)]
	if !ok {
		return "", model.User{}, fmt.Errorf("unsupported provider: %s", providerName)
	}

	token, err := provider.Exchange(ctx, code)
	if err != nil {
		return "", model.User{}, fmt.Errorf("exchanging code: %w", err)
	}

	info, err := provider.FetchUser(ctx, token)
	if err != nil {
		return "", model.User{}, fmt.Errorf("fetching user profile: %w", err)
	}

	user, err := s.userStore.GetOrCreateByEmail(ctx, model.User{
		Email:       info.Email,
		DisplayName: info.DisplayName,
		AvatarURL:   info.AvatarURL,
		Provider:    model.AuthProvider(providerName),
		ProviderID:  info.ProviderID,
	})
	if err != nil {
		return "", model.User{}, fmt.Errorf("upserting user: %w", err)
	}

	jwt, err := auth.SignToken(user.ID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return "", model.User{}, fmt.Errorf("signing JWT: %w", err)
	}

	return jwt, user, nil
}

// UpdatePreferredLanguage validates and updates the user's preferred language.
func (s *AuthService) UpdatePreferredLanguage(ctx context.Context, userID model.UserID, lang string) (model.User, error) {
	if !SupportedLanguages[lang] {
		return model.User{}, model.ErrUnsupportedLanguage
	}
	user, err := s.userStore.UpdatePreferredLanguage(ctx, userID, lang)
	if err != nil {
		return model.User{}, fmt.Errorf("updating preferred language: %w", err)
	}
	return user, nil
}

// GetUser returns a user by ID.
func (s *AuthService) GetUser(ctx context.Context, userID model.UserID) (model.User, error) {
	return s.userStore.GetByID(ctx, userID)
}
