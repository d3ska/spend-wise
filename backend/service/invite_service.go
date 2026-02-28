package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"backend/model"
)

// InviteStoreIface defines the invite store methods used by InviteService.
type InviteStoreIface interface {
	Create(ctx context.Context, inv model.Invite) (model.Invite, error)
	GetByCode(ctx context.Context, code string) (model.Invite, string, string, error)
	MarkUsed(ctx context.Context, inviteID int64, userID model.UserID) error
}

// InviteService handles workspace invite operations.
type InviteService struct {
	inviteStore    InviteStoreIface
	workspaceStore WorkspaceStoreIface
	defaultExpiry  time.Duration
}

// NewInviteService creates a new InviteService.
func NewInviteService(is InviteStoreIface, ws WorkspaceStoreIface, defaultExpiryHours int) *InviteService {
	return &InviteService{
		inviteStore:    is,
		workspaceStore: ws,
		defaultExpiry:  time.Duration(defaultExpiryHours) * time.Hour,
	}
}

// CreateInviteInput holds input for creating an invite.
type CreateInviteInput struct {
	Role           model.MemberRole
	ExpiresInHours int // 0 means use default
}

// CreateInvite generates a single-use invite link for a workspace.
// Only owners can create invites. Role must be editor or viewer.
func (s *InviteService) CreateInvite(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, input CreateInviteInput) (model.Invite, error) {
	// Verify caller is owner
	member, err := s.workspaceStore.GetMember(ctx, wsID, userID)
	if err != nil {
		return model.Invite{}, err
	}
	if member.Role != model.RoleOwner {
		return model.Invite{}, model.ErrInsufficientPermission
	}

	// Validate role — only editor or viewer allowed
	if input.Role != model.RoleEditor && input.Role != model.RoleViewer {
		return model.Invite{}, model.ErrInvalidInviteRole
	}

	code, err := generateInviteCode()
	if err != nil {
		return model.Invite{}, fmt.Errorf("generating invite code: %w", err)
	}

	expiry := s.defaultExpiry
	if input.ExpiresInHours > 0 {
		expiry = time.Duration(input.ExpiresInHours) * time.Hour
	}

	inv := model.Invite{
		WorkspaceID: wsID,
		Code:        code,
		Role:        input.Role,
		CreatedBy:   userID,
		ExpiresAt:   time.Now().Add(expiry),
	}

	created, err := s.inviteStore.Create(ctx, inv)
	if err != nil {
		return model.Invite{}, fmt.Errorf("creating invite: %w", err)
	}
	return created, nil
}

// PreviewInvite returns preview information about an invite without accepting it.
// This is a public endpoint (no auth required).
func (s *InviteService) PreviewInvite(ctx context.Context, code string) (model.InvitePreview, error) {
	inv, workspaceName, inviterName, err := s.inviteStore.GetByCode(ctx, code)
	if err != nil {
		return model.InvitePreview{}, err
	}

	preview := model.InvitePreview{
		WorkspaceName: workspaceName,
		InviterName:   inviterName,
		Role:          inv.Role,
		ExpiresAt:     inv.ExpiresAt,
		Expired:       time.Now().After(inv.ExpiresAt),
		Used:          inv.UsedBy != nil,
	}
	return preview, nil
}

// AcceptInvite accepts an invite and adds the user to the workspace.
func (s *InviteService) AcceptInvite(ctx context.Context, userID model.UserID, code string) (model.Workspace, error) {
	inv, _, _, err := s.inviteStore.GetByCode(ctx, code)
	if err != nil {
		return model.Workspace{}, err
	}

	// Check if expired
	if time.Now().After(inv.ExpiresAt) {
		return model.Workspace{}, model.ErrInviteExpired
	}

	// Check if already used
	if inv.UsedBy != nil {
		return model.Workspace{}, model.ErrInviteUsed
	}

	// Check if already a member
	_, err = s.workspaceStore.GetMember(ctx, inv.WorkspaceID, userID)
	if err == nil {
		return model.Workspace{}, model.ErrAlreadyMember
	}
	if err != model.ErrNotWorkspaceMember {
		return model.Workspace{}, fmt.Errorf("checking membership: %w", err)
	}

	// Add member
	if _, err := s.workspaceStore.AddMember(ctx, inv.WorkspaceID, userID, inv.Role); err != nil {
		return model.Workspace{}, fmt.Errorf("adding member: %w", err)
	}

	// Mark invite as used
	if err := s.inviteStore.MarkUsed(ctx, inv.ID, userID); err != nil {
		return model.Workspace{}, fmt.Errorf("marking invite used: %w", err)
	}

	// Return the workspace
	ws, err := s.workspaceStore.GetByID(ctx, inv.WorkspaceID)
	if err != nil {
		return model.Workspace{}, fmt.Errorf("getting workspace: %w", err)
	}
	return ws, nil
}

func generateInviteCode() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
