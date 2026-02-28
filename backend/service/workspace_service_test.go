package service

import (
	"context"
	"errors"
	"testing"

	"backend/model"
)

// ── Mock stores ──

type mockWorkspaceStore struct {
	CreateFunc                  func(ctx context.Context, ws model.Workspace) (model.Workspace, error)
	GetByIDFunc                 func(ctx context.Context, id model.WorkspaceID) (model.Workspace, error)
	ListByUserFunc              func(ctx context.Context, userID model.UserID) ([]model.Workspace, error)
	UpdateFunc                  func(ctx context.Context, id model.WorkspaceID, name, description string) (model.Workspace, error)
	DeleteFunc                  func(ctx context.Context, id model.WorkspaceID) error
	AddMemberFunc               func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) (model.WorkspaceMember, error)
	GetMemberFunc               func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error)
	ListMembersFunc             func(ctx context.Context, wsID model.WorkspaceID) ([]model.WorkspaceMember, error)
	ListMembersWithProfilesFunc func(ctx context.Context, wsID model.WorkspaceID) ([]model.MemberWithProfile, error)
	UpdateMemberRoleFunc        func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) error
	RemoveMemberFunc            func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) error
}

func (m *mockWorkspaceStore) Create(ctx context.Context, ws model.Workspace) (model.Workspace, error) {
	return m.CreateFunc(ctx, ws)
}
func (m *mockWorkspaceStore) GetByID(ctx context.Context, id model.WorkspaceID) (model.Workspace, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockWorkspaceStore) ListByUser(ctx context.Context, userID model.UserID) ([]model.Workspace, error) {
	return m.ListByUserFunc(ctx, userID)
}
func (m *mockWorkspaceStore) Update(ctx context.Context, id model.WorkspaceID, name, description string) (model.Workspace, error) {
	return m.UpdateFunc(ctx, id, name, description)
}
func (m *mockWorkspaceStore) Delete(ctx context.Context, id model.WorkspaceID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *mockWorkspaceStore) AddMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) (model.WorkspaceMember, error) {
	return m.AddMemberFunc(ctx, wsID, userID, role)
}
func (m *mockWorkspaceStore) GetMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
	return m.GetMemberFunc(ctx, wsID, userID)
}
func (m *mockWorkspaceStore) ListMembers(ctx context.Context, wsID model.WorkspaceID) ([]model.WorkspaceMember, error) {
	return m.ListMembersFunc(ctx, wsID)
}
func (m *mockWorkspaceStore) ListMembersWithProfiles(ctx context.Context, wsID model.WorkspaceID) ([]model.MemberWithProfile, error) {
	if m.ListMembersWithProfilesFunc != nil {
		return m.ListMembersWithProfilesFunc(ctx, wsID)
	}
	return nil, nil
}
func (m *mockWorkspaceStore) UpdateMemberRole(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) error {
	if m.UpdateMemberRoleFunc != nil {
		return m.UpdateMemberRoleFunc(ctx, wsID, userID, role)
	}
	return nil
}
func (m *mockWorkspaceStore) RemoveMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) error {
	return m.RemoveMemberFunc(ctx, wsID, userID)
}

type mockCategoryStore struct {
	CreateFunc          func(ctx context.Context, cat model.Category) (model.Category, error)
	GetByIDFunc         func(ctx context.Context, id model.CategoryID) (model.Category, error)
	ListByWorkspaceFunc func(ctx context.Context, wsID model.WorkspaceID) ([]model.Category, error)
	UpdateFunc          func(ctx context.Context, id model.CategoryID, name, icon string) (model.Category, error)
	GetByNameFunc       func(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error)
	GetBySlugFunc       func(ctx context.Context, wsID model.WorkspaceID, slug string) (model.Category, error)
	DeleteFunc          func(ctx context.Context, id model.CategoryID) error
}

func (m *mockCategoryStore) Create(ctx context.Context, cat model.Category) (model.Category, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, cat)
	}
	return model.Category{}, nil
}
func (m *mockCategoryStore) GetByID(ctx context.Context, id model.CategoryID) (model.Category, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockCategoryStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]model.Category, error) {
	return m.ListByWorkspaceFunc(ctx, wsID)
}
func (m *mockCategoryStore) Update(ctx context.Context, id model.CategoryID, name, icon string) (model.Category, error) {
	return m.UpdateFunc(ctx, id, name, icon)
}
func (m *mockCategoryStore) GetByName(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, wsID, name)
	}
	return model.Category{}, model.ErrCategoryNotFound
}
func (m *mockCategoryStore) GetBySlug(ctx context.Context, wsID model.WorkspaceID, slug string) (model.Category, error) {
	if m.GetBySlugFunc != nil {
		return m.GetBySlugFunc(ctx, wsID, slug)
	}
	return model.Category{}, model.ErrCategoryNotFound
}
func (m *mockCategoryStore) Delete(ctx context.Context, id model.CategoryID) error {
	return m.DeleteFunc(ctx, id)
}

// ── Helpers ──

func memberWithRole(role model.MemberRole) model.WorkspaceMember {
	return model.WorkspaceMember{WorkspaceID: 1, UserID: 1, Role: role}
}

// ── Tests ──

func TestWorkspaceService_CreateWorkspace(t *testing.T) {
	tests := map[string]struct {
		input   CreateWorkspaceInput
		member  model.WorkspaceMember
		wantErr error
	}{
		"valid input": {
			input: CreateWorkspaceInput{Name: "Home", Description: "Home budget"},
		},
		"empty name": {
			input:   CreateWorkspaceInput{Name: "", Description: "desc"},
			wantErr: model.ErrWorkspaceNameRequired,
		},
		"empty description": {
			input:   CreateWorkspaceInput{Name: "Home", Description: ""},
			wantErr: model.ErrWorkspaceDescriptionRequired,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				CreateFunc: func(_ context.Context, w model.Workspace) (model.Workspace, error) {
					w.ID = 1
					return w, nil
				},
				AddMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID, _ model.MemberRole) (model.WorkspaceMember, error) {
					return memberWithRole(model.RoleOwner), nil
				},
			}
			cs := &mockCategoryStore{
				CreateFunc: func(_ context.Context, cat model.Category) (model.Category, error) {
					cat.ID = 1
					return cat, nil
				},
			}
			svc := NewWorkspaceService(ws, cs)

			got, err := svc.CreateWorkspace(context.Background(), 1, tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tc.input.Name {
				t.Errorf("got name %q, want %q", got.Name, tc.input.Name)
			}
		})
	}
}

func TestWorkspaceService_GetWorkspace(t *testing.T) {
	tests := map[string]struct {
		memberErr error
		member    model.WorkspaceMember
		wantErr   error
	}{
		"viewer allowed": {
			member: memberWithRole(model.RoleViewer),
		},
		"non-member": {
			memberErr: model.ErrNotWorkspaceMember,
			wantErr:   model.ErrNotWorkspaceMember,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					if tc.memberErr != nil {
						return model.WorkspaceMember{}, tc.memberErr
					}
					return tc.member, nil
				},
				GetByIDFunc: func(_ context.Context, _ model.WorkspaceID) (model.Workspace, error) {
					return model.Workspace{ID: 1, Name: "Home"}, nil
				},
			}
			svc := NewWorkspaceService(ws, &mockCategoryStore{})

			_, err := svc.GetWorkspace(context.Background(), 1, 1)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkspaceService_UpdateWorkspace(t *testing.T) {
	tests := map[string]struct {
		role    model.MemberRole
		input   UpdateWorkspaceInput
		wantErr error
	}{
		"editor allowed": {
			role:  model.RoleEditor,
			input: UpdateWorkspaceInput{Name: "Updated", Description: "New desc"},
		},
		"viewer rejected": {
			role:    model.RoleViewer,
			input:   UpdateWorkspaceInput{Name: "Updated", Description: "New desc"},
			wantErr: model.ErrInsufficientPermission,
		},
		"validation failure": {
			role:    model.RoleEditor,
			input:   UpdateWorkspaceInput{Name: "", Description: "New desc"},
			wantErr: model.ErrWorkspaceNameRequired,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
				UpdateFunc: func(_ context.Context, _ model.WorkspaceID, n, d string) (model.Workspace, error) {
					return model.Workspace{ID: 1, Name: n, Description: d}, nil
				},
			}
			svc := NewWorkspaceService(ws, &mockCategoryStore{})

			_, err := svc.UpdateWorkspace(context.Background(), 1, 1, tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkspaceService_DeleteWorkspace(t *testing.T) {
	tests := map[string]struct {
		role    model.MemberRole
		wantErr error
	}{
		"owner allowed": {
			role: model.RoleOwner,
		},
		"editor rejected": {
			role:    model.RoleEditor,
			wantErr: model.ErrInsufficientPermission,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
				DeleteFunc: func(_ context.Context, _ model.WorkspaceID) error {
					return nil
				},
			}
			svc := NewWorkspaceService(ws, &mockCategoryStore{})

			err := svc.DeleteWorkspace(context.Background(), 1, 1)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkspaceService_AddMember(t *testing.T) {
	tests := map[string]struct {
		role    model.MemberRole
		wantErr error
	}{
		"owner allowed": {
			role: model.RoleOwner,
		},
		"editor rejected": {
			role:    model.RoleEditor,
			wantErr: model.ErrInsufficientPermission,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
				AddMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID, _ model.MemberRole) (model.WorkspaceMember, error) {
					return memberWithRole(model.RoleEditor), nil
				},
			}
			svc := NewWorkspaceService(ws, &mockCategoryStore{})

			_, err := svc.AddMember(context.Background(), 1, 1, AddMemberInput{UserID: 2, Role: model.RoleEditor})
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkspaceService_RemoveMember(t *testing.T) {
	tests := map[string]struct {
		role    model.MemberRole
		wantErr error
	}{
		"owner allowed": {
			role: model.RoleOwner,
		},
		"editor rejected": {
			role:    model.RoleEditor,
			wantErr: model.ErrInsufficientPermission,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
				RemoveMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) error {
					return nil
				},
			}
			svc := NewWorkspaceService(ws, &mockCategoryStore{})

			err := svc.RemoveMember(context.Background(), 1, 1, 2)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkspaceService_CreateCategory(t *testing.T) {
	tests := map[string]struct {
		role    model.MemberRole
		wantErr error
	}{
		"editor allowed": {
			role: model.RoleEditor,
		},
		"viewer rejected": {
			role:    model.RoleViewer,
			wantErr: model.ErrInsufficientPermission,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
			}
			cs := &mockCategoryStore{
				CreateFunc: func(_ context.Context, c model.Category) (model.Category, error) {
					c.ID = 1
					return c, nil
				},
			}
			svc := NewWorkspaceService(ws, cs)

			_, err := svc.CreateCategory(context.Background(), 1, 1, CreateCategoryInput{Name: "Food", Icon: "🍔"})
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkspaceService_DeleteCategory(t *testing.T) {
	uncatSlug := "uncategorized"
	tests := map[string]struct {
		role    model.MemberRole
		catSlug *string
		wantErr error
	}{
		"editor deletes normal category": {
			role:    model.RoleEditor,
			catSlug: nil,
		},
		"viewer rejected": {
			role:    model.RoleViewer,
			catSlug: nil,
			wantErr: model.ErrInsufficientPermission,
		},
		"Uncategorized is protected": {
			role:    model.RoleEditor,
			catSlug: &uncatSlug,
			wantErr: model.ErrCategoryUndeletable,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
			}
			cs := &mockCategoryStore{
				GetByIDFunc: func(_ context.Context, _ model.CategoryID) (model.Category, error) {
					return model.Category{ID: 10, WorkspaceID: 1, Name: "Food", Slug: tc.catSlug}, nil
				},
				DeleteFunc: func(_ context.Context, _ model.CategoryID) error {
					return nil
				},
			}
			svc := NewWorkspaceService(ws, cs)

			err := svc.DeleteCategory(context.Background(), 1, 1, 10)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkspaceService_ListMembersWithProfiles(t *testing.T) {
	tests := map[string]struct {
		role    model.MemberRole
		wantErr error
	}{
		"viewer allowed": {
			role: model.RoleViewer,
		},
		"non-member rejected": {
			role:    model.MemberRole(""),
			wantErr: model.ErrNotWorkspaceMember,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					if tc.wantErr != nil {
						return model.WorkspaceMember{}, tc.wantErr
					}
					return memberWithRole(tc.role), nil
				},
				ListMembersWithProfilesFunc: func(_ context.Context, _ model.WorkspaceID) ([]model.MemberWithProfile, error) {
					return []model.MemberWithProfile{
						{UserID: 1, DisplayName: "Alice", Role: model.RoleOwner},
						{UserID: 2, DisplayName: "Bob", Role: model.RoleEditor},
					}, nil
				},
			}
			svc := NewWorkspaceService(ws, &mockCategoryStore{})

			members, err := svc.ListMembersWithProfiles(context.Background(), 1, 1)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(members) != 2 {
				t.Errorf("got %d members, want 2", len(members))
			}
		})
	}
}

func TestWorkspaceService_UpdateMemberRole(t *testing.T) {
	tests := map[string]struct {
		callerRole   model.MemberRole
		targetUserID model.UserID
		newRole      model.MemberRole
		wantErr      error
	}{
		"owner changes editor to viewer": {
			callerRole:   model.RoleOwner,
			targetUserID: 2,
			newRole:      model.RoleViewer,
		},
		"owner changes viewer to editor": {
			callerRole:   model.RoleOwner,
			targetUserID: 2,
			newRole:      model.RoleEditor,
		},
		"editor rejected": {
			callerRole:   model.RoleEditor,
			targetUserID: 2,
			newRole:      model.RoleViewer,
			wantErr:      model.ErrInsufficientPermission,
		},
		"cannot change own role": {
			callerRole:   model.RoleOwner,
			targetUserID: 1, // same as caller
			newRole:      model.RoleEditor,
			wantErr:      model.ErrCannotChangeOwnRole,
		},
		"cannot set owner role": {
			callerRole:   model.RoleOwner,
			targetUserID: 2,
			newRole:      model.RoleOwner,
			wantErr:      model.ErrInsufficientPermission,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.callerRole), nil
				},
				UpdateMemberRoleFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID, _ model.MemberRole) error {
					return nil
				},
			}
			svc := NewWorkspaceService(ws, &mockCategoryStore{})

			err := svc.UpdateMemberRole(context.Background(), 1, 1, tc.targetUserID, tc.newRole)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
