package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/auth"
	"backend/model"

	"golang.org/x/oauth2"
)

// ── Mock user store ──

type mockUserStore struct {
	GetOrCreateByEmailFunc      func(ctx context.Context, u model.User) (model.User, error)
	GetByIDFunc                 func(ctx context.Context, id model.UserID) (model.User, error)
	UpdatePreferredLanguageFunc func(ctx context.Context, id model.UserID, lang string) (model.User, error)
}

func (m *mockUserStore) GetOrCreateByEmail(ctx context.Context, u model.User) (model.User, error) {
	return m.GetOrCreateByEmailFunc(ctx, u)
}

func (m *mockUserStore) GetByID(ctx context.Context, id model.UserID) (model.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return model.User{}, model.ErrUserNotFound
}

func (m *mockUserStore) UpdatePreferredLanguage(ctx context.Context, id model.UserID, lang string) (model.User, error) {
	if m.UpdatePreferredLanguageFunc != nil {
		return m.UpdatePreferredLanguageFunc(ctx, id, lang)
	}
	return model.User{}, model.ErrUserNotFound
}

// ── Mock auth provider ──

type mockAuthProvider struct {
	AuthCodeURLFunc string
	ExchangeFunc    func(ctx context.Context, code string) (*oauth2.Token, error)
	FetchUserFunc   func(ctx context.Context, token *oauth2.Token) (auth.UserInfo, error)
}

func (m *mockAuthProvider) AuthCodeURL(state string) string {
	return m.AuthCodeURLFunc
}

func (m *mockAuthProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return m.ExchangeFunc(ctx, code)
}

func (m *mockAuthProvider) FetchUser(ctx context.Context, token *oauth2.Token) (auth.UserInfo, error) {
	return m.FetchUserFunc(ctx, token)
}

// ── Tests ──

func TestAuthService_Authenticate(t *testing.T) {
	tests := map[string]struct {
		provider    string
		exchangeErr error
		fetchErr    error
		wantErr     bool
	}{
		"success": {
			provider: "google",
		},
		"unsupported provider": {
			provider: "unknown",
			wantErr:  true,
		},
		"exchange failure": {
			provider:    "google",
			exchangeErr: errors.New("exchange failed"),
			wantErr:     true,
		},
		"fetch user failure": {
			provider: "google",
			fetchErr: errors.New("fetch failed"),
			wantErr:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockProvider := &mockAuthProvider{
				ExchangeFunc: func(_ context.Context, _ string) (*oauth2.Token, error) {
					if tc.exchangeErr != nil {
						return nil, tc.exchangeErr
					}
					return &oauth2.Token{AccessToken: "test-token"}, nil
				},
				FetchUserFunc: func(_ context.Context, _ *oauth2.Token) (auth.UserInfo, error) {
					if tc.fetchErr != nil {
						return auth.UserInfo{}, tc.fetchErr
					}
					return auth.UserInfo{
						Email:       "test@example.com",
						DisplayName: "Test User",
						AvatarURL:   "https://example.com/avatar.jpg",
						ProviderID:  "google-123",
					}, nil
				},
			}

			us := &mockUserStore{
				GetOrCreateByEmailFunc: func(_ context.Context, u model.User) (model.User, error) {
					u.ID = 1
					return u, nil
				},
			}

			providers := map[model.AuthProvider]auth.Provider{
				"google": mockProvider,
			}

			svc := NewAuthService(providers, us, "test-secret", 24*time.Hour)

			jwt, user, err := svc.Authenticate(context.Background(), tc.provider, "test-code")
			if tc.wantErr {
				if err == nil {
					t.Fatal("got nil error, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if jwt == "" {
				t.Error("got empty JWT, want non-empty")
			}
			if user.Email != "test@example.com" {
				t.Errorf("got email %q, want %q", user.Email, "test@example.com")
			}
			if user.ID != 1 {
				t.Errorf("got user ID %v, want 1", user.ID)
			}
		})
	}
}

func TestAuthService_Authenticate_CreatesUser(t *testing.T) {
	var capturedUser model.User

	mockProvider := &mockAuthProvider{
		ExchangeFunc: func(_ context.Context, _ string) (*oauth2.Token, error) {
			return &oauth2.Token{AccessToken: "test-token"}, nil
		},
		FetchUserFunc: func(_ context.Context, _ *oauth2.Token) (auth.UserInfo, error) {
			return auth.UserInfo{
				Email:       "alice@example.com",
				DisplayName: "Alice",
				AvatarURL:   "https://example.com/alice.jpg",
				ProviderID:  "google-456",
			}, nil
		},
	}

	us := &mockUserStore{
		GetOrCreateByEmailFunc: func(_ context.Context, u model.User) (model.User, error) {
			capturedUser = u
			u.ID = 42
			return u, nil
		},
	}

	providers := map[model.AuthProvider]auth.Provider{
		"google": mockProvider,
	}

	svc := NewAuthService(providers, us, "test-secret", 24*time.Hour)

	_, user, err := svc.Authenticate(context.Background(), "google", "test-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify user data was correctly passed to store
	if capturedUser.Email != "alice@example.com" {
		t.Errorf("got email %q, want %q", capturedUser.Email, "alice@example.com")
	}
	if capturedUser.DisplayName != "Alice" {
		t.Errorf("got display name %q, want %q", capturedUser.DisplayName, "Alice")
	}
	if capturedUser.Provider != "google" {
		t.Errorf("got provider %q, want %q", capturedUser.Provider, "google")
	}
	if capturedUser.ProviderID != "google-456" {
		t.Errorf("got provider ID %q, want %q", capturedUser.ProviderID, "google-456")
	}

	// Verify returned user has ID from store
	if user.ID != 42 {
		t.Errorf("got user ID %v, want 42", user.ID)
	}
}

func TestAuthService_UpdatePreferredLanguage(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		storeErr error
		wantErr  error
		wantLang string
	}{
		{
			name:     "valid language en",
			lang:     "en",
			wantLang: "en",
		},
		{
			name:     "valid language pl",
			lang:     "pl",
			wantLang: "pl",
		},
		{
			name:    "unsupported language",
			lang:    "fr",
			wantErr: model.ErrUnsupportedLanguage,
		},
		{
			name:    "empty language",
			lang:    "",
			wantErr: model.ErrUnsupportedLanguage,
		},
		{
			name:     "user not found",
			lang:     "pl",
			storeErr: model.ErrUserNotFound,
			wantErr:  model.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &mockUserStore{
				GetOrCreateByEmailFunc: func(_ context.Context, u model.User) (model.User, error) {
					return u, nil
				},
				UpdatePreferredLanguageFunc: func(_ context.Context, id model.UserID, lang string) (model.User, error) {
					if tt.storeErr != nil {
						return model.User{}, tt.storeErr
					}
					return model.User{ID: id, PreferredLanguage: lang}, nil
				},
			}

			svc := NewAuthService(nil, us, "test-secret", 24*time.Hour)

			user, err := svc.UpdatePreferredLanguage(context.Background(), model.UserID(1), tt.lang)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.PreferredLanguage != tt.wantLang {
				t.Errorf("got language %q, want %q", user.PreferredLanguage, tt.wantLang)
			}
		})
	}
}
