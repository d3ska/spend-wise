package model

import (
	"errors"
	"testing"
)

func TestWorkspace_Validate(t *testing.T) {
	tests := map[string]struct {
		ws      Workspace
		wantErr error
	}{
		"valid": {
			ws:      Workspace{Name: "Home", Description: "Home budget"},
			wantErr: nil,
		},
		"empty name": {
			ws:      Workspace{Name: "", Description: "desc"},
			wantErr: ErrWorkspaceNameRequired,
		},
		"empty description": {
			ws:      Workspace{Name: "Home", Description: ""},
			wantErr: ErrWorkspaceDescriptionRequired,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.ws.Validate()
			if tc.wantErr == nil {
				if err != nil {
					t.Errorf("got %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestMemberRole_Level(t *testing.T) {
	tests := map[string]struct {
		role MemberRole
		want int
	}{
		"owner":   {role: RoleOwner, want: 3},
		"editor":  {role: RoleEditor, want: 2},
		"viewer":  {role: RoleViewer, want: 1},
		"unknown": {role: MemberRole("invalid"), want: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.role.Level()
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestWorkspaceMember_CanEdit(t *testing.T) {
	tests := map[string]struct {
		role     MemberRole
		wantEdit bool
		wantView bool
	}{
		"owner":   {role: RoleOwner, wantEdit: true, wantView: true},
		"editor":  {role: RoleEditor, wantEdit: true, wantView: true},
		"viewer":  {role: RoleViewer, wantEdit: false, wantView: true},
		"unknown": {role: MemberRole("unknown"), wantEdit: false, wantView: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := WorkspaceMember{Role: tc.role}

			if got := m.CanEdit(); got != tc.wantEdit {
				t.Errorf("CanEdit() got %v, want %v", got, tc.wantEdit)
			}
			if got := m.CanView(); got != tc.wantView {
				t.Errorf("CanView() got %v, want %v", got, tc.wantView)
			}
		})
	}
}
