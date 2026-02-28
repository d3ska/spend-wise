package store

import (
	"context"
	"errors"
	"fmt"

	"backend/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserStore provides user persistence operations.
type UserStore struct {
	q *Queries
}

// NewUserStore creates a new UserStore.
func NewUserStore(db DBTX) *UserStore {
	return &UserStore{q: New(db)}
}

// GetByID returns a user by ID. Returns model.ErrUserNotFound if not found.
func (s *UserStore) GetByID(ctx context.Context, id model.UserID) (model.User, error) {
	row, err := s.q.GetUserByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("getting user by id: %w", err)
	}
	return rowToModelUser(row.ID, row.Email, row.DisplayName, row.AvatarUrl, row.Provider, row.ProviderID, row.PreferredLanguage, row.CreatedAt, row.UpdatedAt), nil
}

// GetOrCreateByEmail looks up a user by email. If found, updates provider info
// and returns the existing user. If not found, creates a new user.
func (s *UserStore) GetOrCreateByEmail(ctx context.Context, u model.User) (model.User, error) {
	existing, err := s.q.GetUserByEmail(ctx, u.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, fmt.Errorf("looking up user by email: %w", err)
	}

	if err == nil {
		// User exists — update provider info.
		updated, err := s.q.UpdateUserProvider(ctx, UpdateUserProviderParams{
			ID:          existing.ID,
			DisplayName: u.DisplayName,
			AvatarUrl:   u.AvatarURL,
			Provider:    AuthProvider(u.Provider),
			ProviderID:  u.ProviderID,
		})
		if err != nil {
			return model.User{}, fmt.Errorf("updating user provider: %w", err)
		}
		return rowToModelUser(updated.ID, updated.Email, updated.DisplayName, updated.AvatarUrl, updated.Provider, updated.ProviderID, updated.PreferredLanguage, updated.CreatedAt, updated.UpdatedAt), nil
	}

	// User does not exist — create.
	created, err := s.q.InsertUser(ctx, InsertUserParams{
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarUrl:   u.AvatarURL,
		Provider:    AuthProvider(u.Provider),
		ProviderID:  u.ProviderID,
	})
	if err != nil {
		return model.User{}, fmt.Errorf("inserting user: %w", err)
	}
	return rowToModelUser(created.ID, created.Email, created.DisplayName, created.AvatarUrl, created.Provider, created.ProviderID, created.PreferredLanguage, created.CreatedAt, created.UpdatedAt), nil
}

// UpdatePreferredLanguage updates a user's preferred language.
// Returns model.ErrUserNotFound if the user does not exist.
func (s *UserStore) UpdatePreferredLanguage(ctx context.Context, id model.UserID, lang string) (model.User, error) {
	row, err := s.q.UpdatePreferredLanguage(ctx, UpdatePreferredLanguageParams{
		ID:                int64(id),
		PreferredLanguage: lang,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("updating preferred language: %w", err)
	}
	return rowToModelUser(row.ID, row.Email, row.DisplayName, row.AvatarUrl, row.Provider, row.ProviderID, row.PreferredLanguage, row.CreatedAt, row.UpdatedAt), nil
}

func rowToModelUser(id int64, email, displayName, avatarURL string, provider AuthProvider, providerID, preferredLanguage string, createdAt, updatedAt pgtype.Timestamptz) model.User {
	return model.User{
		ID:                model.UserID(id),
		Email:             email,
		DisplayName:       displayName,
		AvatarURL:         avatarURL,
		Provider:          model.AuthProvider(provider),
		ProviderID:        providerID,
		PreferredLanguage: preferredLanguage,
		CreatedAt:         createdAt.Time,
		UpdatedAt:         updatedAt.Time,
	}
}
