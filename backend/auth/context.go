package auth

import (
	"context"

	"backend/model"
)

type contextKey int

const userIDKey contextKey = iota

// WithUserID stores a UserID in the context.
func WithUserID(ctx context.Context, id model.UserID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserIDFromContext retrieves the UserID from the request context.
// Returns 0 and false if not present.
func UserIDFromContext(ctx context.Context) (model.UserID, bool) {
	id, ok := ctx.Value(userIDKey).(model.UserID)
	return id, ok
}
