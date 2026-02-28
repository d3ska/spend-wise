package auth

import (
	"context"

	"golang.org/x/oauth2"
)

// UserInfo holds the profile data fetched from an OAuth provider.
type UserInfo struct {
	Email       string
	DisplayName string
	AvatarURL   string
	ProviderID  string
}

// Provider defines the interface for an OAuth2 authentication provider.
type Provider interface {
	// AuthCodeURL returns the URL to redirect the user to for authentication.
	AuthCodeURL(state string) string

	// Exchange trades an authorization code for an access token.
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)

	// FetchUser retrieves the user's profile using the access token.
	FetchUser(ctx context.Context, token *oauth2.Token) (UserInfo, error)
}
