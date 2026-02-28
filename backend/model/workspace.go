package model

import (
	"strings"
	"time"
)

// WorkspaceID is a typed wrapper for workspace identifiers.
type WorkspaceID int64

// Workspace represents a top-level container for financial data.
type Workspace struct {
	ID          WorkspaceID
	Name        string
	Description string
	OwnerID     UserID
	MemberCount int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Invite represents a workspace invite link.
type Invite struct {
	ID          int64
	WorkspaceID WorkspaceID
	Code        string
	Role        MemberRole
	CreatedBy   UserID
	ExpiresAt   time.Time
	UsedBy      *UserID
	UsedAt      *time.Time
	CreatedAt   time.Time
}

// InvitePreview holds preview data for an invite, shown before accepting.
type InvitePreview struct {
	WorkspaceName string
	InviterName   string
	Role          MemberRole
	ExpiresAt     time.Time
	Expired       bool
	Used          bool
}

// MemberWithProfile is a workspace member enriched with user profile data.
type MemberWithProfile struct {
	WorkspaceID WorkspaceID
	UserID      UserID
	Role        MemberRole
	CreatedAt   time.Time
	DisplayName string
	Email       string
	AvatarURL   string
}

// Workspace field limits.
const (
	MaxWorkspaceNameLen        = 100
	MaxWorkspaceDescriptionLen = 500
)

// Validate returns an error if the workspace has invalid fields.
func (w Workspace) Validate() error {
	name := strings.TrimSpace(w.Name)
	if name == "" {
		return ErrWorkspaceNameRequired
	}
	if len(name) > MaxWorkspaceNameLen {
		return ErrWorkspaceNameTooLong
	}
	desc := strings.TrimSpace(w.Description)
	if desc == "" {
		return ErrWorkspaceDescriptionRequired
	}
	if len(desc) > MaxWorkspaceDescriptionLen {
		return ErrWorkspaceDescriptionTooLong
	}
	return nil
}

// MemberRole represents a workspace member's role.
type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleEditor MemberRole = "editor"
	RoleViewer MemberRole = "viewer"
)

// Level returns a numeric level for role comparison. Higher = more permissions.
func (r MemberRole) Level() int {
	switch r {
	case RoleOwner:
		return 3
	case RoleEditor:
		return 2
	case RoleViewer:
		return 1
	default:
		return 0
	}
}

// WorkspaceMember represents a user's membership in a workspace.
type WorkspaceMember struct {
	WorkspaceID WorkspaceID
	UserID      UserID
	Role        MemberRole
	CreatedAt   time.Time
}

// CanEdit returns true if the member has editor or owner permissions.
func (m WorkspaceMember) CanEdit() bool {
	return m.Role.Level() >= RoleEditor.Level()
}

// CanView returns true if the member has any role (viewer, editor, or owner).
func (m WorkspaceMember) CanView() bool {
	return m.Role.Level() >= RoleViewer.Level()
}
