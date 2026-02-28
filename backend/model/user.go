package model

import "time"

// UserID is a typed wrapper for user identifiers.
type UserID int64

// AuthProvider represents an OAuth2 authentication provider.
type AuthProvider string

const (
	ProviderGoogle AuthProvider = "google"
	ProviderGitHub AuthProvider = "github"
)

// User represents an authenticated user, created via OAuth2 SSO.
type User struct {
	ID                UserID
	Email             string
	DisplayName       string
	AvatarURL         string
	Provider          AuthProvider
	ProviderID        string
	PreferredLanguage string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
