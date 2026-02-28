package service

import (
	"context"
	"fmt"
	"log/slog"

	"backend/model"
)

// defaultCategory holds the definition for a default category seeded on workspace creation.
type defaultCategory struct {
	Slug string
	Name string
	Icon string
}

// defaultCategories is the list of categories seeded into every new workspace.
var defaultCategories = []defaultCategory{
	{Slug: "groceries", Name: "Groceries", Icon: "🛒"},
	{Slug: "transportation", Name: "Transportation", Icon: "🚌"},
	{Slug: "automotive", Name: "Automotive", Icon: "🚗"},
	{Slug: "shopping", Name: "Shopping", Icon: "🛍️"},
	{Slug: "subscriptions", Name: "Subscriptions", Icon: "🔄"},
	{Slug: "gifts", Name: "Gifts", Icon: "🎁"},
	{Slug: "dining", Name: "Dining", Icon: "🍽️"},
	{Slug: "utilities", Name: "Utilities & Rent", Icon: "🏠"},
	{Slug: "debt", Name: "Debt", Icon: "💳"},
	{Slug: "supplements", Name: "Supplements", Icon: "💊"},
	{Slug: "uncategorized", Name: "Uncategorized", Icon: "📂"},
}

// WorkspaceStoreIface defines the workspace store methods used by WorkspaceService.
type WorkspaceStoreIface interface {
	Create(ctx context.Context, ws model.Workspace) (model.Workspace, error)
	GetByID(ctx context.Context, id model.WorkspaceID) (model.Workspace, error)
	ListByUser(ctx context.Context, userID model.UserID) ([]model.Workspace, error)
	Update(ctx context.Context, id model.WorkspaceID, name, description string) (model.Workspace, error)
	Delete(ctx context.Context, id model.WorkspaceID) error
	AddMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) (model.WorkspaceMember, error)
	GetMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error)
	ListMembers(ctx context.Context, wsID model.WorkspaceID) ([]model.WorkspaceMember, error)
	ListMembersWithProfiles(ctx context.Context, wsID model.WorkspaceID) ([]model.MemberWithProfile, error)
	UpdateMemberRole(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) error
	RemoveMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) error
}

// CategoryStoreIface defines the category store methods used by services.
type CategoryStoreIface interface {
	Create(ctx context.Context, cat model.Category) (model.Category, error)
	GetByID(ctx context.Context, id model.CategoryID) (model.Category, error)
	ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]model.Category, error)
	Update(ctx context.Context, id model.CategoryID, name, icon string) (model.Category, error)
	GetByName(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error)
	GetBySlug(ctx context.Context, wsID model.WorkspaceID, slug string) (model.Category, error)
	Delete(ctx context.Context, id model.CategoryID) error
}

// CreateWorkspaceInput holds input for creating a workspace.
type CreateWorkspaceInput struct {
	Name        string
	Description string
}

// UpdateWorkspaceInput holds input for updating a workspace.
type UpdateWorkspaceInput struct {
	Name        string
	Description string
}

// AddMemberInput holds input for adding a member to a workspace.
type AddMemberInput struct {
	UserID model.UserID
	Role   model.MemberRole
}

// CreateCategoryInput holds input for creating a category.
type CreateCategoryInput struct {
	Name string
	Icon string
}

// UpdateCategoryInput holds input for updating a category.
type UpdateCategoryInput struct {
	Name string
	Icon string
}

// WorkspaceService handles workspace, member, and category operations
// with role-based permission checks.
type WorkspaceService struct {
	workspaceStore WorkspaceStoreIface
	categoryStore  CategoryStoreIface
}

// NewWorkspaceService creates a new WorkspaceService.
func NewWorkspaceService(ws WorkspaceStoreIface, cs CategoryStoreIface) *WorkspaceService {
	return &WorkspaceService{
		workspaceStore: ws,
		categoryStore:  cs,
	}
}

// requireMembership checks that a user is a member of the workspace
// with at least the given minimum role. Returns the member or an error.
func (s *WorkspaceService) requireMembership(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, minRole model.MemberRole) (model.WorkspaceMember, error) {
	member, err := s.workspaceStore.GetMember(ctx, wsID, userID)
	if err != nil {
		return model.WorkspaceMember{}, err
	}
	if member.Role.Level() < minRole.Level() {
		return model.WorkspaceMember{}, model.ErrInsufficientPermission
	}
	return member, nil
}

// ── Workspace CRUD ──

// CreateWorkspace validates input, persists the workspace, and auto-adds
// the creator as owner.
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, userID model.UserID, input CreateWorkspaceInput) (model.Workspace, error) {
	ws := model.Workspace{
		Name:        input.Name,
		Description: input.Description,
		OwnerID:     userID,
	}
	if err := ws.Validate(); err != nil {
		return model.Workspace{}, err
	}

	created, err := s.workspaceStore.Create(ctx, ws)
	if err != nil {
		return model.Workspace{}, fmt.Errorf("creating workspace: %w", err)
	}

	if _, err := s.workspaceStore.AddMember(ctx, created.ID, userID, model.RoleOwner); err != nil {
		return model.Workspace{}, fmt.Errorf("adding owner as member: %w", err)
	}

	s.seedDefaultCategories(ctx, created.ID)

	return created, nil
}

// seedDefaultCategories creates the predefined default categories for a new workspace.
func (s *WorkspaceService) seedDefaultCategories(ctx context.Context, wsID model.WorkspaceID) {
	for _, dc := range defaultCategories {
		slug := dc.Slug
		_, err := s.categoryStore.Create(ctx, model.Category{
			WorkspaceID: wsID,
			Name:        dc.Name,
			Icon:        dc.Icon,
			Slug:        &slug,
		})
		if err != nil {
			slog.Error("failed to seed default category", "slug", dc.Slug, "workspace_id", wsID, "error", err)
		}
	}
}

// GetWorkspace returns a workspace if the user is a member (viewer+).
func (s *WorkspaceService) GetWorkspace(ctx context.Context, userID model.UserID, wsID model.WorkspaceID) (model.Workspace, error) {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleViewer); err != nil {
		return model.Workspace{}, err
	}
	return s.workspaceStore.GetByID(ctx, wsID)
}

// ListWorkspaces returns all workspaces where the user is a member.
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID model.UserID) ([]model.Workspace, error) {
	return s.workspaceStore.ListByUser(ctx, userID)
}

// UpdateWorkspace updates a workspace. Requires editor+ role.
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, input UpdateWorkspaceInput) (model.Workspace, error) {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return model.Workspace{}, err
	}

	ws := model.Workspace{Name: input.Name, Description: input.Description}
	if err := ws.Validate(); err != nil {
		return model.Workspace{}, err
	}

	return s.workspaceStore.Update(ctx, wsID, input.Name, input.Description)
}

// DeleteWorkspace deletes a workspace. Requires owner role.
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, userID model.UserID, wsID model.WorkspaceID) error {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleOwner); err != nil {
		return err
	}
	return s.workspaceStore.Delete(ctx, wsID)
}

// ── Member Management ──

// ListMembersWithProfiles returns all members with profile data. Requires viewer+ role.
func (s *WorkspaceService) ListMembersWithProfiles(ctx context.Context, userID model.UserID, wsID model.WorkspaceID) ([]model.MemberWithProfile, error) {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleViewer); err != nil {
		return nil, err
	}
	return s.workspaceStore.ListMembersWithProfiles(ctx, wsID)
}

// UpdateMemberRole changes a member's role. Requires owner role.
// Cannot change own role or set role to owner.
func (s *WorkspaceService) UpdateMemberRole(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, targetUserID model.UserID, newRole model.MemberRole) error {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleOwner); err != nil {
		return err
	}
	if userID == targetUserID {
		return model.ErrCannotChangeOwnRole
	}
	if newRole == model.RoleOwner {
		return model.ErrInsufficientPermission
	}
	return s.workspaceStore.UpdateMemberRole(ctx, wsID, targetUserID, newRole)
}

// AddMember adds a user to a workspace. Requires owner role.
// The assigned role must be editor or viewer -- cannot grant owner role.
func (s *WorkspaceService) AddMember(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, input AddMemberInput) (model.WorkspaceMember, error) {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleOwner); err != nil {
		return model.WorkspaceMember{}, err
	}
	// Prevent granting owner role via AddMember (owner is only set at workspace creation).
	if input.Role == model.RoleOwner {
		return model.WorkspaceMember{}, model.ErrInsufficientPermission
	}
	return s.workspaceStore.AddMember(ctx, wsID, input.UserID, input.Role)
}

// RemoveMember removes a user from a workspace. Requires owner role.
// The owner cannot remove themselves (use DeleteWorkspace instead).
func (s *WorkspaceService) RemoveMember(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, targetUserID model.UserID) error {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleOwner); err != nil {
		return err
	}
	// Prevent owner from removing themselves -- this would orphan the workspace.
	if userID == targetUserID {
		return model.ErrCannotRemoveSelf
	}
	return s.workspaceStore.RemoveMember(ctx, wsID, targetUserID)
}

// ── Category Operations ──

// CreateCategory creates a category in a workspace. Requires editor+ role.
func (s *WorkspaceService) CreateCategory(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, input CreateCategoryInput) (model.Category, error) {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return model.Category{}, err
	}
	cat := model.Category{
		WorkspaceID: wsID,
		Name:        input.Name,
		Icon:        input.Icon,
	}
	if err := cat.Validate(); err != nil {
		return model.Category{}, err
	}
	return s.categoryStore.Create(ctx, cat)
}

// ListCategories returns all categories in a workspace. Requires viewer+ role.
func (s *WorkspaceService) ListCategories(ctx context.Context, userID model.UserID, wsID model.WorkspaceID) ([]model.Category, error) {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleViewer); err != nil {
		return nil, err
	}
	return s.categoryStore.ListByWorkspace(ctx, wsID)
}

// UpdateCategory updates a category. Requires editor+ role.
// Verifies the category belongs to the target workspace to prevent cross-workspace IDOR.
func (s *WorkspaceService) UpdateCategory(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, catID model.CategoryID, input UpdateCategoryInput) (model.Category, error) {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return model.Category{}, err
	}

	// Verify category belongs to this workspace (prevents cross-workspace IDOR).
	existing, err := s.categoryStore.GetByID(ctx, catID)
	if err != nil {
		return model.Category{}, err
	}
	if existing.WorkspaceID != wsID {
		return model.Category{}, model.ErrCategoryNotFound
	}

	cat := model.Category{Name: input.Name, Icon: input.Icon}
	if err := cat.Validate(); err != nil {
		return model.Category{}, err
	}
	return s.categoryStore.Update(ctx, catID, input.Name, input.Icon)
}

// DeleteCategory deletes a category. Requires editor+ role.
// The "Uncategorized" category cannot be deleted.
// Verifies the category belongs to the target workspace to prevent cross-workspace IDOR.
func (s *WorkspaceService) DeleteCategory(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, catID model.CategoryID) error {
	if _, err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return err
	}

	cat, err := s.categoryStore.GetByID(ctx, catID)
	if err != nil {
		return err
	}
	// Verify category belongs to this workspace (prevents cross-workspace IDOR).
	if cat.WorkspaceID != wsID {
		return model.ErrCategoryNotFound
	}
	if cat.Slug != nil && *cat.Slug == "uncategorized" {
		return model.ErrCategoryUndeletable
	}

	return s.categoryStore.Delete(ctx, catID)
}
