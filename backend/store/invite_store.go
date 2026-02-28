package store

import (
	"context"
	"errors"
	"fmt"

	"backend/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// InviteStore provides invite persistence operations.
type InviteStore struct {
	q *Queries
}

// NewInviteStore creates a new InviteStore.
func NewInviteStore(db DBTX) *InviteStore {
	return &InviteStore{q: New(db)}
}

// Create persists a new invite and returns it.
func (s *InviteStore) Create(ctx context.Context, inv model.Invite) (model.Invite, error) {
	row, err := s.q.InsertInvite(ctx, InsertInviteParams{
		WorkspaceID: int64(inv.WorkspaceID),
		Code:        inv.Code,
		Role:        MemberRole(inv.Role),
		CreatedBy:   int64(inv.CreatedBy),
		ExpiresAt:   pgtype.Timestamptz{Time: inv.ExpiresAt, Valid: true},
	})
	if err != nil {
		return model.Invite{}, fmt.Errorf("inserting invite: %w", err)
	}
	return toModelInvite(row), nil
}

// GetByCode returns an invite by its code with workspace and inviter info.
// Returns model.ErrInviteNotFound if not found.
func (s *InviteStore) GetByCode(ctx context.Context, code string) (model.Invite, string, string, error) {
	row, err := s.q.GetInviteByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Invite{}, "", "", model.ErrInviteNotFound
		}
		return model.Invite{}, "", "", fmt.Errorf("getting invite by code: %w", err)
	}
	inv := model.Invite{
		ID:          row.ID,
		WorkspaceID: model.WorkspaceID(row.WorkspaceID),
		Code:        row.Code,
		Role:        model.MemberRole(row.Role),
		CreatedBy:   model.UserID(row.CreatedBy),
		ExpiresAt:   row.ExpiresAt.Time,
		CreatedAt:   row.CreatedAt.Time,
	}
	if row.UsedBy.Valid {
		uid := model.UserID(row.UsedBy.Int64)
		inv.UsedBy = &uid
	}
	if row.UsedAt.Valid {
		inv.UsedAt = &row.UsedAt.Time
	}
	return inv, row.WorkspaceName, row.InviterName, nil
}

// MarkUsed marks an invite as used by a specific user.
func (s *InviteStore) MarkUsed(ctx context.Context, inviteID int64, userID model.UserID) error {
	if err := s.q.MarkInviteUsed(ctx, MarkInviteUsedParams{
		ID:     inviteID,
		UsedBy: pgtype.Int8{Int64: int64(userID), Valid: true},
	}); err != nil {
		return fmt.Errorf("marking invite used: %w", err)
	}
	return nil
}

func toModelInvite(row WorkspaceInvite) model.Invite {
	inv := model.Invite{
		ID:          row.ID,
		WorkspaceID: model.WorkspaceID(row.WorkspaceID),
		Code:        row.Code,
		Role:        model.MemberRole(row.Role),
		CreatedBy:   model.UserID(row.CreatedBy),
		ExpiresAt:   row.ExpiresAt.Time,
		CreatedAt:   row.CreatedAt.Time,
	}
	if row.UsedBy.Valid {
		uid := model.UserID(row.UsedBy.Int64)
		inv.UsedBy = &uid
	}
	if row.UsedAt.Valid {
		inv.UsedAt = &row.UsedAt.Time
	}
	return inv
}
