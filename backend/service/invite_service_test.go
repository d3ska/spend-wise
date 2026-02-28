package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/model"
)

// ── Mock invite store ──

type mockInviteStore struct {
	CreateFunc    func(ctx context.Context, inv model.Invite) (model.Invite, error)
	GetByCodeFunc func(ctx context.Context, code string) (model.Invite, string, string, error)
	MarkUsedFunc  func(ctx context.Context, inviteID int64, userID model.UserID) error
}

func (m *mockInviteStore) Create(ctx context.Context, inv model.Invite) (model.Invite, error) {
	return m.CreateFunc(ctx, inv)
}
func (m *mockInviteStore) GetByCode(ctx context.Context, code string) (model.Invite, string, string, error) {
	return m.GetByCodeFunc(ctx, code)
}
func (m *mockInviteStore) MarkUsed(ctx context.Context, inviteID int64, userID model.UserID) error {
	return m.MarkUsedFunc(ctx, inviteID, userID)
}

// ── Tests ──

func TestInviteService_CreateInvite(t *testing.T) {
	tests := map[string]struct {
		callerRole model.MemberRole
		input      CreateInviteInput
		wantErr    error
	}{
		"owner creates editor invite": {
			callerRole: model.RoleOwner,
			input:      CreateInviteInput{Role: model.RoleEditor},
		},
		"owner creates viewer invite": {
			callerRole: model.RoleOwner,
			input:      CreateInviteInput{Role: model.RoleViewer},
		},
		"editor rejected": {
			callerRole: model.RoleEditor,
			input:      CreateInviteInput{Role: model.RoleViewer},
			wantErr:    model.ErrInsufficientPermission,
		},
		"viewer rejected": {
			callerRole: model.RoleViewer,
			input:      CreateInviteInput{Role: model.RoleViewer},
			wantErr:    model.ErrInsufficientPermission,
		},
		"invalid role owner": {
			callerRole: model.RoleOwner,
			input:      CreateInviteInput{Role: model.RoleOwner},
			wantErr:    model.ErrInvalidInviteRole,
		},
		"invalid role empty": {
			callerRole: model.RoleOwner,
			input:      CreateInviteInput{Role: ""},
			wantErr:    model.ErrInvalidInviteRole,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.callerRole), nil
				},
			}
			is := &mockInviteStore{
				CreateFunc: func(_ context.Context, inv model.Invite) (model.Invite, error) {
					inv.ID = 1
					return inv, nil
				},
			}
			svc := NewInviteService(is, ws, 168)

			got, err := svc.CreateInvite(context.Background(), 1, 1, tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Code == "" {
				t.Error("expected non-empty invite code")
			}
			if got.Role != tc.input.Role {
				t.Errorf("got role %q, want %q", got.Role, tc.input.Role)
			}
		})
	}
}

func TestInviteService_PreviewInvite(t *testing.T) {
	tests := map[string]struct {
		setupInvite func() (model.Invite, string, string, error)
		wantExpired bool
		wantUsed    bool
		wantErr     error
	}{
		"valid invite": {
			setupInvite: func() (model.Invite, string, string, error) {
				return model.Invite{
					ID:        1,
					Role:      model.RoleEditor,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, "Home Budget", "Alice", nil
			},
		},
		"expired invite": {
			setupInvite: func() (model.Invite, string, string, error) {
				return model.Invite{
					ID:        1,
					Role:      model.RoleEditor,
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}, "Home Budget", "Alice", nil
			},
			wantExpired: true,
		},
		"used invite": {
			setupInvite: func() (model.Invite, string, string, error) {
				uid := model.UserID(2)
				return model.Invite{
					ID:        1,
					Role:      model.RoleEditor,
					ExpiresAt: time.Now().Add(24 * time.Hour),
					UsedBy:    &uid,
				}, "Home Budget", "Alice", nil
			},
			wantUsed: true,
		},
		"not found": {
			setupInvite: func() (model.Invite, string, string, error) {
				return model.Invite{}, "", "", model.ErrInviteNotFound
			},
			wantErr: model.ErrInviteNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			is := &mockInviteStore{
				GetByCodeFunc: func(_ context.Context, _ string) (model.Invite, string, string, error) {
					return tc.setupInvite()
				},
			}
			svc := NewInviteService(is, &mockWorkspaceStore{}, 168)

			preview, err := svc.PreviewInvite(context.Background(), "abc123")
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if preview.Expired != tc.wantExpired {
				t.Errorf("got expired %v, want %v", preview.Expired, tc.wantExpired)
			}
			if preview.Used != tc.wantUsed {
				t.Errorf("got used %v, want %v", preview.Used, tc.wantUsed)
			}
		})
	}
}

func TestInviteService_AcceptInvite(t *testing.T) {
	validInvite := model.Invite{
		ID:          1,
		WorkspaceID: 10,
		Code:        "abc123",
		Role:        model.RoleEditor,
		CreatedBy:   1,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	tests := map[string]struct {
		invite       model.Invite
		getMemberErr error
		wantErr      error
	}{
		"valid accept": {
			invite:       validInvite,
			getMemberErr: model.ErrNotWorkspaceMember,
		},
		"expired invite": {
			invite: model.Invite{
				ID:        1,
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			wantErr: model.ErrInviteExpired,
		},
		"used invite": {
			invite: func() model.Invite {
				uid := model.UserID(3)
				return model.Invite{
					ID:        1,
					ExpiresAt: time.Now().Add(24 * time.Hour),
					UsedBy:    &uid,
				}
			}(),
			wantErr: model.ErrInviteUsed,
		},
		"already member": {
			invite:       validInvite,
			getMemberErr: nil, // no error means user IS a member
			wantErr:      model.ErrAlreadyMember,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			is := &mockInviteStore{
				GetByCodeFunc: func(_ context.Context, _ string) (model.Invite, string, string, error) {
					return tc.invite, "Home", "Alice", nil
				},
				MarkUsedFunc: func(_ context.Context, _ int64, _ model.UserID) error {
					return nil
				},
			}
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					if tc.getMemberErr != nil {
						return model.WorkspaceMember{}, tc.getMemberErr
					}
					return memberWithRole(model.RoleEditor), nil
				},
				AddMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID, _ model.MemberRole) (model.WorkspaceMember, error) {
					return memberWithRole(model.RoleEditor), nil
				},
				GetByIDFunc: func(_ context.Context, id model.WorkspaceID) (model.Workspace, error) {
					return model.Workspace{ID: id, Name: "Home"}, nil
				},
			}
			svc := NewInviteService(is, ws, 168)

			_, err := svc.AcceptInvite(context.Background(), 2, "abc123")
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
