package store

import (
	"context"
	"errors"
	"fmt"

	"backend/model"

	"github.com/jackc/pgx/v5"
)

// WorkspaceStore provides workspace and member persistence operations.
type WorkspaceStore struct {
	q *Queries
}

// NewWorkspaceStore creates a new WorkspaceStore.
func NewWorkspaceStore(db DBTX) *WorkspaceStore {
	return &WorkspaceStore{q: New(db)}
}

// Create persists a new workspace and returns it with a generated ID.
func (s *WorkspaceStore) Create(ctx context.Context, ws model.Workspace) (model.Workspace, error) {
	row, err := s.q.InsertWorkspace(ctx, InsertWorkspaceParams{
		Name:        ws.Name,
		Description: ws.Description,
		OwnerID:     int64(ws.OwnerID),
	})
	if err != nil {
		return model.Workspace{}, fmt.Errorf("inserting workspace: %w", err)
	}
	return toModelWorkspace(row), nil
}

// GetByID returns a workspace by ID. Returns model.ErrWorkspaceNotFound if not found.
func (s *WorkspaceStore) GetByID(ctx context.Context, id model.WorkspaceID) (model.Workspace, error) {
	row, err := s.q.GetWorkspaceByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Workspace{}, model.ErrWorkspaceNotFound
		}
		return model.Workspace{}, fmt.Errorf("getting workspace by id: %w", err)
	}
	return toModelWorkspace(row), nil
}

// ListByUser returns all workspaces where the user is a member, including member counts.
func (s *WorkspaceStore) ListByUser(ctx context.Context, userID model.UserID) ([]model.Workspace, error) {
	rows, err := s.q.ListWorkspacesByUser(ctx, int64(userID))
	if err != nil {
		return nil, fmt.Errorf("listing workspaces by user: %w", err)
	}
	result := make([]model.Workspace, 0, len(rows))
	for _, row := range rows {
		result = append(result, model.Workspace{
			ID:          model.WorkspaceID(row.ID),
			Name:        row.Name,
			Description: row.Description,
			OwnerID:     model.UserID(row.OwnerID),
			MemberCount: row.MemberCount,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return result, nil
}

// Update updates a workspace's name and description.
func (s *WorkspaceStore) Update(ctx context.Context, id model.WorkspaceID, name, description string) (model.Workspace, error) {
	row, err := s.q.UpdateWorkspace(ctx, UpdateWorkspaceParams{
		ID:          int64(id),
		Name:        name,
		Description: description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Workspace{}, model.ErrWorkspaceNotFound
		}
		return model.Workspace{}, fmt.Errorf("updating workspace: %w", err)
	}
	return toModelWorkspace(row), nil
}

// Delete removes a workspace by ID. Cascades to members and categories.
func (s *WorkspaceStore) Delete(ctx context.Context, id model.WorkspaceID) error {
	if err := s.q.DeleteWorkspace(ctx, int64(id)); err != nil {
		return fmt.Errorf("deleting workspace: %w", err)
	}
	return nil
}

// AddMember adds a user to a workspace with the given role.
func (s *WorkspaceStore) AddMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) (model.WorkspaceMember, error) {
	row, err := s.q.InsertMember(ctx, InsertMemberParams{
		WorkspaceID: int64(wsID),
		UserID:      int64(userID),
		Role:        MemberRole(role),
	})
	if err != nil {
		return model.WorkspaceMember{}, fmt.Errorf("inserting member: %w", err)
	}
	return toModelMember(row), nil
}

// GetMember returns a workspace member. Returns model.ErrNotWorkspaceMember if not found.
func (s *WorkspaceStore) GetMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
	row, err := s.q.GetMember(ctx, GetMemberParams{
		WorkspaceID: int64(wsID),
		UserID:      int64(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
		}
		return model.WorkspaceMember{}, fmt.Errorf("getting member: %w", err)
	}
	return toModelMember(row), nil
}

// ListMembers returns all members of a workspace.
func (s *WorkspaceStore) ListMembers(ctx context.Context, wsID model.WorkspaceID) ([]model.WorkspaceMember, error) {
	rows, err := s.q.ListMembers(ctx, int64(wsID))
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	result := make([]model.WorkspaceMember, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelMember(row))
	}
	return result, nil
}

// RemoveMember removes a user from a workspace.
func (s *WorkspaceStore) RemoveMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) error {
	if err := s.q.DeleteMember(ctx, DeleteMemberParams{
		WorkspaceID: int64(wsID),
		UserID:      int64(userID),
	}); err != nil {
		return fmt.Errorf("deleting member: %w", err)
	}
	return nil
}

// ListMembersWithProfiles returns all members of a workspace with their user profile data.
func (s *WorkspaceStore) ListMembersWithProfiles(ctx context.Context, wsID model.WorkspaceID) ([]model.MemberWithProfile, error) {
	rows, err := s.q.ListMembersWithProfiles(ctx, int64(wsID))
	if err != nil {
		return nil, fmt.Errorf("listing members with profiles: %w", err)
	}
	result := make([]model.MemberWithProfile, 0, len(rows))
	for _, row := range rows {
		result = append(result, model.MemberWithProfile{
			WorkspaceID: model.WorkspaceID(row.WorkspaceID),
			UserID:      model.UserID(row.UserID),
			Role:        model.MemberRole(row.Role),
			CreatedAt:   row.CreatedAt.Time,
			DisplayName: row.DisplayName,
			Email:       row.Email,
			AvatarURL:   row.AvatarUrl,
		})
	}
	return result, nil
}

// UpdateMemberRole updates a member's role in a workspace.
func (s *WorkspaceStore) UpdateMemberRole(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) error {
	if err := s.q.UpdateMemberRole(ctx, UpdateMemberRoleParams{
		WorkspaceID: int64(wsID),
		UserID:      int64(userID),
		Role:        MemberRole(role),
	}); err != nil {
		return fmt.Errorf("updating member role: %w", err)
	}
	return nil
}

func toModelWorkspace(row Workspace) model.Workspace {
	return model.Workspace{
		ID:          model.WorkspaceID(row.ID),
		Name:        row.Name,
		Description: row.Description,
		OwnerID:     model.UserID(row.OwnerID),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func toModelMember(row WorkspaceMember) model.WorkspaceMember {
	return model.WorkspaceMember{
		WorkspaceID: model.WorkspaceID(row.WorkspaceID),
		UserID:      model.UserID(row.UserID),
		Role:        model.MemberRole(row.Role),
		CreatedAt:   row.CreatedAt.Time,
	}
}
