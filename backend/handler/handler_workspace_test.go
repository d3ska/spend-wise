package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/model"
	"backend/service"
)

func TestHandleServiceError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "ErrWorkspaceNotFound returns 404",
			err:        model.ErrWorkspaceNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrCategoryNotFound returns 404",
			err:        model.ErrCategoryNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrFundingNotFound returns 404",
			err:        model.ErrFundingNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrRuleNotFound returns 404",
			err:        model.ErrRuleNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrBankConnectionNotFound returns 404",
			err:        model.ErrBankConnectionNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrBankAccountNotFound returns 404",
			err:        model.ErrBankAccountNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrNotWorkspaceMember returns 403",
			err:        model.ErrNotWorkspaceMember,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "ErrInsufficientPermission returns 403",
			err:        model.ErrInsufficientPermission,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "ErrWorkspaceNameRequired returns 400",
			err:        model.ErrWorkspaceNameRequired,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrWorkspaceDescriptionRequired returns 400",
			err:        model.ErrWorkspaceDescriptionRequired,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrCategoryBudgetsExceed returns 400",
			err:        model.ErrCategoryBudgetsExceed,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrCategoryUndeletable returns 400",
			err:        model.ErrCategoryUndeletable,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown error returns 500",
			err:        errors.New("unknown error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handleServiceError(w, tt.err)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}

			var env codedErrorEnvelope
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Fatalf("decoding error envelope: %v", err)
			}

			if env.Message == "" {
				t.Error("want non-empty error message")
			}
		})
	}
}

// Mock workspace store for testing
type mockWorkspaceStore struct {
	createFunc                  func(ctx context.Context, ws model.Workspace) (model.Workspace, error)
	getByIDFunc                 func(ctx context.Context, id model.WorkspaceID) (model.Workspace, error)
	listByUserFunc              func(ctx context.Context, userID model.UserID) ([]model.Workspace, error)
	updateFunc                  func(ctx context.Context, id model.WorkspaceID, name, description string) (model.Workspace, error)
	deleteFunc                  func(ctx context.Context, id model.WorkspaceID) error
	addMemberFunc               func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) (model.WorkspaceMember, error)
	getMemberFunc               func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error)
	listMembersFunc             func(ctx context.Context, wsID model.WorkspaceID) ([]model.WorkspaceMember, error)
	listMembersWithProfilesFunc func(ctx context.Context, wsID model.WorkspaceID) ([]model.MemberWithProfile, error)
	updateMemberRoleFunc        func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) error
	removeMemberFunc            func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) error
}

func (m *mockWorkspaceStore) Create(ctx context.Context, ws model.Workspace) (model.Workspace, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, ws)
	}
	return model.Workspace{}, nil
}

func (m *mockWorkspaceStore) GetByID(ctx context.Context, id model.WorkspaceID) (model.Workspace, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return model.Workspace{}, nil
}

func (m *mockWorkspaceStore) ListByUser(ctx context.Context, userID model.UserID) ([]model.Workspace, error) {
	if m.listByUserFunc != nil {
		return m.listByUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockWorkspaceStore) Update(ctx context.Context, id model.WorkspaceID, name, description string) (model.Workspace, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, name, description)
	}
	return model.Workspace{}, nil
}

func (m *mockWorkspaceStore) Delete(ctx context.Context, id model.WorkspaceID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockWorkspaceStore) AddMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) (model.WorkspaceMember, error) {
	if m.addMemberFunc != nil {
		return m.addMemberFunc(ctx, wsID, userID, role)
	}
	return model.WorkspaceMember{}, nil
}

func (m *mockWorkspaceStore) GetMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
	if m.getMemberFunc != nil {
		return m.getMemberFunc(ctx, wsID, userID)
	}
	return model.WorkspaceMember{}, nil
}

func (m *mockWorkspaceStore) ListMembers(ctx context.Context, wsID model.WorkspaceID) ([]model.WorkspaceMember, error) {
	if m.listMembersFunc != nil {
		return m.listMembersFunc(ctx, wsID)
	}
	return nil, nil
}

func (m *mockWorkspaceStore) ListMembersWithProfiles(ctx context.Context, wsID model.WorkspaceID) ([]model.MemberWithProfile, error) {
	if m.listMembersWithProfilesFunc != nil {
		return m.listMembersWithProfilesFunc(ctx, wsID)
	}
	return nil, nil
}

func (m *mockWorkspaceStore) UpdateMemberRole(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) error {
	if m.updateMemberRoleFunc != nil {
		return m.updateMemberRoleFunc(ctx, wsID, userID, role)
	}
	return nil
}

func (m *mockWorkspaceStore) RemoveMember(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) error {
	if m.removeMemberFunc != nil {
		return m.removeMemberFunc(ctx, wsID, userID)
	}
	return nil
}

// Mock category store for testing
type mockCategoryStore struct {
	createFunc          func(ctx context.Context, cat model.Category) (model.Category, error)
	getByIDFunc         func(ctx context.Context, id model.CategoryID) (model.Category, error)
	listByWorkspaceFunc func(ctx context.Context, wsID model.WorkspaceID) ([]model.Category, error)
	updateFunc          func(ctx context.Context, id model.CategoryID, name, icon string) (model.Category, error)
	getByNameFunc       func(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error)
	getBySlugFunc       func(ctx context.Context, wsID model.WorkspaceID, slug string) (model.Category, error)
	deleteFunc          func(ctx context.Context, id model.CategoryID) error
}

func (m *mockCategoryStore) Create(ctx context.Context, cat model.Category) (model.Category, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, cat)
	}
	return model.Category{}, nil
}

func (m *mockCategoryStore) GetByID(ctx context.Context, id model.CategoryID) (model.Category, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return model.Category{}, nil
}

func (m *mockCategoryStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]model.Category, error) {
	if m.listByWorkspaceFunc != nil {
		return m.listByWorkspaceFunc(ctx, wsID)
	}
	return nil, nil
}

func (m *mockCategoryStore) Update(ctx context.Context, id model.CategoryID, name, icon string) (model.Category, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, name, icon)
	}
	return model.Category{}, nil
}

func (m *mockCategoryStore) GetByName(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(ctx, wsID, name)
	}
	return model.Category{}, nil
}

func (m *mockCategoryStore) GetBySlug(ctx context.Context, wsID model.WorkspaceID, slug string) (model.Category, error) {
	if m.getBySlugFunc != nil {
		return m.getBySlugFunc(ctx, wsID, slug)
	}
	return model.Category{}, model.ErrCategoryNotFound
}

func (m *mockCategoryStore) Delete(ctx context.Context, id model.CategoryID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestWorkspaceHandler_CreateWorkspace(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		body       string
		setupStore func(*mockWorkspaceStore, *mockCategoryStore)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success returns 201",
			userID: 1,
			body:   `{"name":"My Workspace","description":"Test workspace"}`,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {
				ws.createFunc = func(ctx context.Context, w model.Workspace) (model.Workspace, error) {
					return model.Workspace{
						ID:          1,
						Name:        w.Name,
						Description: w.Description,
						OwnerID:     w.OwnerID,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil
				}
				ws.addMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, role model.MemberRole) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        role,
						CreatedAt:   time.Now(),
					}, nil
				}
				cs.createFunc = func(ctx context.Context, cat model.Category) (model.Category, error) {
					cat.ID = 1
					return cat, nil
				}
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			body:       `{"name":"Test","description":"Test"}`,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "invalid body returns 400",
			userID:     1,
			body:       `{invalid json}`,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			cs := &mockCategoryStore{}
			tt.setupStore(ws, cs)

			svc := service.NewWorkspaceService(ws, cs)
			h := NewWorkspaceHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", bytes.NewBufferString(tt.body))
			} else {
				r = newAuthenticatedRequest(t, http.MethodPost, "/api/v1/workspaces", bytes.NewBufferString(tt.body), tt.userID)
			}

			w := httptest.NewRecorder()
			h.CreateWorkspace(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}

			if !tt.wantErr {
				var resp map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if resp["name"] != "My Workspace" {
					t.Errorf("want name 'My Workspace', got %v", resp["name"])
				}
			}
		})
	}
}

func TestWorkspaceHandler_GetWorkspace(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		wsID       string
		setupStore func(*mockWorkspaceStore, *mockCategoryStore)
		wantStatus int
	}{
		{
			name:   "success returns 200",
			userID: 1,
			wsID:   "1",
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleOwner,
					}, nil
				}
				ws.getByIDFunc = func(ctx context.Context, id model.WorkspaceID) (model.Workspace, error) {
					return model.Workspace{
						ID:          id,
						Name:        "Test Workspace",
						Description: "Description",
						OwnerID:     1,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "not found returns 404",
			userID: 1,
			wsID:   "999",
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			cs := &mockCategoryStore{}
			tt.setupStore(ws, cs)

			svc := service.NewWorkspaceService(ws, cs)
			h := NewWorkspaceHandler(svc, NewSSEHub())

			r := newAuthenticatedRequest(t, http.MethodGet, "/api/v1/workspaces/"+tt.wsID, nil, tt.userID)
			r = withChiParam(r, "id", tt.wsID)

			w := httptest.NewRecorder()
			h.GetWorkspace(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestWorkspaceHandler_DeleteWorkspace(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		wsID       string
		setupStore func(*mockWorkspaceStore, *mockCategoryStore)
		wantStatus int
	}{
		{
			name:   "success returns 204",
			userID: 1,
			wsID:   "1",
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleOwner,
					}, nil
				}
				ws.deleteFunc = func(ctx context.Context, id model.WorkspaceID) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			cs := &mockCategoryStore{}
			tt.setupStore(ws, cs)

			svc := service.NewWorkspaceService(ws, cs)
			h := NewWorkspaceHandler(svc, NewSSEHub())

			r := newAuthenticatedRequest(t, http.MethodDelete, "/api/v1/workspaces/"+tt.wsID, nil, tt.userID)
			r = withChiParam(r, "id", tt.wsID)

			w := httptest.NewRecorder()
			h.DeleteWorkspace(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestWorkspaceHandler_ListWorkspaces(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		setupStore func(*mockWorkspaceStore, *mockCategoryStore)
		wantStatus int
		wantCount  int
	}{
		{
			name:   "returns JSON array with 200",
			userID: 1,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {
				ws.listByUserFunc = func(ctx context.Context, userID model.UserID) ([]model.Workspace, error) {
					return []model.Workspace{
						{
							ID:          1,
							Name:        "Workspace 1",
							Description: "First workspace",
							OwnerID:     userID,
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						},
						{
							ID:          2,
							Name:        "Workspace 2",
							Description: "Second workspace",
							OwnerID:     userID,
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			cs := &mockCategoryStore{}
			tt.setupStore(ws, cs)

			svc := service.NewWorkspaceService(ws, cs)
			h := NewWorkspaceHandler(svc, NewSSEHub())

			r := newAuthenticatedRequest(t, http.MethodGet, "/api/v1/workspaces", nil, tt.userID)

			w := httptest.NewRecorder()
			h.ListWorkspaces(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}

			var resp []map[string]any
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decoding response: %v", err)
			}

			if len(resp) != tt.wantCount {
				t.Errorf("want %d workspaces, got %d", tt.wantCount, len(resp))
			}
		})
	}
}
