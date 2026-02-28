package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/model"
	"backend/service"

	"github.com/shopspring/decimal"
)

// ── Mock stores ──

// mockFundingStore implements service.FundingStoreIface for testing.
type mockFundingStore struct {
	upsertFunc          func(ctx context.Context, f model.Funding) (model.Funding, error)
	listByWorkspaceFunc func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
	deleteFunc          func(ctx context.Context, id model.FundingID, wsID model.WorkspaceID) error
}

func (m *mockFundingStore) Upsert(ctx context.Context, f model.Funding) (model.Funding, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, f)
	}
	return model.Funding{}, nil
}

func (m *mockFundingStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	if m.listByWorkspaceFunc != nil {
		return m.listByWorkspaceFunc(ctx, wsID, fromMonth, toMonth)
	}
	return nil, nil
}

func (m *mockFundingStore) Delete(ctx context.Context, id model.FundingID, wsID model.WorkspaceID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, wsID)
	}
	return nil
}

// mockSummaryStore implements service.SummaryStoreIface for testing.
type mockSummaryStore struct {
	totalSpentFunc         func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (model.Money, error)
	transactionCountFunc   func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (int, error)
	spentByCategoryFunc    func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.CategorySpending, error)
	spentByParticipantFunc func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.ParticipantSpending, error)
	spentByDayFunc         func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.DailySpending, error)
}

func (m *mockSummaryStore) TotalSpent(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (model.Money, error) {
	if m.totalSpentFunc != nil {
		return m.totalSpentFunc(ctx, wsID, from, to)
	}
	return model.Zero("PLN"), nil
}

func (m *mockSummaryStore) TransactionCount(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (int, error) {
	if m.transactionCountFunc != nil {
		return m.transactionCountFunc(ctx, wsID, from, to)
	}
	return 0, nil
}

func (m *mockSummaryStore) SpentByCategory(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.CategorySpending, error) {
	if m.spentByCategoryFunc != nil {
		return m.spentByCategoryFunc(ctx, wsID, from, to)
	}
	return nil, nil
}

func (m *mockSummaryStore) SpentByParticipant(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.ParticipantSpending, error) {
	if m.spentByParticipantFunc != nil {
		return m.spentByParticipantFunc(ctx, wsID, from, to)
	}
	return nil, nil
}

func (m *mockSummaryStore) SpentByDay(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.DailySpending, error) {
	if m.spentByDayFunc != nil {
		return m.spentByDayFunc(ctx, wsID, from, to)
	}
	return []model.DailySpending{}, nil
}

// mockFundingStoreForSummary implements service.FundingStoreForSummaryIface for testing.
type mockFundingStoreForSummary struct {
	listByWorkspaceAndUsersFunc func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
	getOverallBudgetsFunc       func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
	getCategoryBudgetsFunc      func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
}

func (m *mockFundingStoreForSummary) ListByWorkspaceAndUsers(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	if m.listByWorkspaceAndUsersFunc != nil {
		return m.listByWorkspaceAndUsersFunc(ctx, wsID, fromMonth, toMonth)
	}
	return nil, nil
}

func (m *mockFundingStoreForSummary) GetOverallBudgets(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	if m.getOverallBudgetsFunc != nil {
		return m.getOverallBudgetsFunc(ctx, wsID, fromMonth, toMonth)
	}
	return nil, nil
}

func (m *mockFundingStoreForSummary) GetCategoryBudgets(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	if m.getCategoryBudgetsFunc != nil {
		return m.getCategoryBudgetsFunc(ctx, wsID, fromMonth, toMonth)
	}
	return nil, nil
}

// ── CategoryHandler tests ──

func TestCategoryHandler_CreateCategory(t *testing.T) {
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
			body:   `{"name":"Food","icon":"utensils"}`,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleEditor,
					}, nil
				}
				cs.createFunc = func(ctx context.Context, cat model.Category) (model.Category, error) {
					return model.Category{
						ID:          1,
						WorkspaceID: cat.WorkspaceID,
						Name:        cat.Name,
						Icon:        cat.Icon,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil
				}
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			body:       `{"name":"Food","icon":"utensils"}`,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			cs := &mockCategoryStore{}
			tt.setupStore(ws, cs)

			svc := service.NewWorkspaceService(ws, cs)
			h := NewCategoryHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/1/categories", bytes.NewBufferString(tt.body))
			} else {
				r = newAuthenticatedRequest(t, http.MethodPost, "/api/v1/workspaces/1/categories", bytes.NewBufferString(tt.body), tt.userID)
			}
			r = withChiParam(r, "id", "1")

			w := httptest.NewRecorder()
			h.CreateCategory(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}

			if !tt.wantErr {
				var resp map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if resp["name"] != "Food" {
					t.Errorf("got name %v, want 'Food'", resp["name"])
				}
			}
		})
	}
}

func TestCategoryHandler_ListCategories(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		setupStore func(*mockWorkspaceStore, *mockCategoryStore)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success returns 200",
			userID: 1,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleViewer,
					}, nil
				}
				cs.listByWorkspaceFunc = func(ctx context.Context, wsID model.WorkspaceID) ([]model.Category, error) {
					return []model.Category{
						{
							ID:          1,
							WorkspaceID: wsID,
							Name:        "Food",
							Icon:        "utensils",
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			cs := &mockCategoryStore{}
			tt.setupStore(ws, cs)

			svc := service.NewWorkspaceService(ws, cs)
			h := NewCategoryHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/1/categories", nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodGet, "/api/v1/workspaces/1/categories", nil, tt.userID)
			}
			r = withChiParam(r, "id", "1")

			w := httptest.NewRecorder()
			h.ListCategories(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}

			if !tt.wantErr {
				var resp []map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if len(resp) != 1 {
					t.Errorf("got %d categories, want 1", len(resp))
				}
			}
		})
	}
}

func TestCategoryHandler_DeleteCategory(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		setupStore func(*mockWorkspaceStore, *mockCategoryStore)
		wantStatus int
	}{
		{
			name:   "success returns 204",
			userID: 1,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleEditor,
					}, nil
				}
				cs.getByIDFunc = func(ctx context.Context, id model.CategoryID) (model.Category, error) {
					return model.Category{
						ID:          id,
						WorkspaceID: 1,
						Name:        "Food",
						Icon:        "utensils",
					}, nil
				}
				cs.deleteFunc = func(ctx context.Context, id model.CategoryID) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			setupStore: func(ws *mockWorkspaceStore, cs *mockCategoryStore) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			cs := &mockCategoryStore{}
			tt.setupStore(ws, cs)

			svc := service.NewWorkspaceService(ws, cs)
			h := NewCategoryHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/1/categories/1", nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodDelete, "/api/v1/workspaces/1/categories/1", nil, tt.userID)
			}
			r = withChiParams(r, map[string]string{"id": "1", "catID": "1"})

			w := httptest.NewRecorder()
			h.DeleteCategory(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}
		})
	}
}

// ── FundingHandler tests ──

func TestFundingHandler_RecordFunding(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		body       string
		setupStore func(*mockWorkspaceStore, *mockFundingStore)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success returns 200",
			userID: 1,
			body:   `{"year_month":"2025-01","amount":{"amount":"100.00","currency":"PLN"}}`,
			setupStore: func(ws *mockWorkspaceStore, fs *mockFundingStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleEditor,
					}, nil
				}
				fs.listByWorkspaceFunc = func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				}
				fs.upsertFunc = func(ctx context.Context, f model.Funding) (model.Funding, error) {
					return model.Funding{
						ID:          1,
						WorkspaceID: f.WorkspaceID,
						UserID:      f.UserID,
						YearMonth:   f.YearMonth,
						Amount:      f.Amount,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			body:       `{"year_month":"2025-01","amount":{"amount":"100.00","currency":"PLN"}}`,
			setupStore: func(ws *mockWorkspaceStore, fs *mockFundingStore) {},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			fs := &mockFundingStore{}
			tt.setupStore(ws, fs)

			svc := service.NewFundingService(fs, ws)
			h := NewFundingHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/1/fundings", bytes.NewBufferString(tt.body))
			} else {
				r = newAuthenticatedRequest(t, http.MethodPut, "/api/v1/workspaces/1/fundings", bytes.NewBufferString(tt.body), tt.userID)
			}
			r = withChiParam(r, "id", "1")

			w := httptest.NewRecorder()
			h.RecordFunding(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}

			if !tt.wantErr {
				var resp map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if resp["year_month"] != "2025-01" {
					t.Errorf("got year_month %v, want '2025-01'", resp["year_month"])
				}
			}
		})
	}
}

func TestFundingHandler_ListFundings(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		query      string
		setupStore func(*mockWorkspaceStore, *mockFundingStore)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success returns 200",
			userID: 1,
			query:  "?from=2025-01&to=2025-03",
			setupStore: func(ws *mockWorkspaceStore, fs *mockFundingStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleViewer,
					}, nil
				}
				fs.listByWorkspaceFunc = func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
					return []model.Funding{
						{
							ID:          1,
							WorkspaceID: wsID,
							UserID:      1,
							YearMonth:   "2025-01",
							Amount:      model.NewMoney(decimal.NewFromInt(100), "PLN"),
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			query:      "?from=2025-01&to=2025-03",
			setupStore: func(ws *mockWorkspaceStore, fs *mockFundingStore) {},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			fs := &mockFundingStore{}
			tt.setupStore(ws, fs)

			svc := service.NewFundingService(fs, ws)
			h := NewFundingHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/1/fundings"+tt.query, nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodGet, "/api/v1/workspaces/1/fundings"+tt.query, nil, tt.userID)
			}
			r = withChiParam(r, "id", "1")

			w := httptest.NewRecorder()
			h.ListFundings(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}

			if !tt.wantErr {
				var resp []map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if len(resp) != 1 {
					t.Errorf("got %d fundings, want 1", len(resp))
				}
			}
		})
	}
}

func TestFundingHandler_DeleteFunding(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		setupStore func(*mockWorkspaceStore, *mockFundingStore)
		wantStatus int
	}{
		{
			name:   "success returns 204",
			userID: 1,
			setupStore: func(ws *mockWorkspaceStore, fs *mockFundingStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleEditor,
					}, nil
				}
				fs.deleteFunc = func(ctx context.Context, id model.FundingID, wsID model.WorkspaceID) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			setupStore: func(ws *mockWorkspaceStore, fs *mockFundingStore) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			fs := &mockFundingStore{}
			tt.setupStore(ws, fs)

			svc := service.NewFundingService(fs, ws)
			h := NewFundingHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/1/fundings/1", nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodDelete, "/api/v1/workspaces/1/fundings/1", nil, tt.userID)
			}
			r = withChiParams(r, map[string]string{"id": "1", "fundingId": "1"})

			w := httptest.NewRecorder()
			h.DeleteFunding(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}
		})
	}
}

// ── RuleHandler tests ──

func TestRuleHandler_CreateRule(t *testing.T) {
	wsID := model.WorkspaceID(1)

	tests := []struct {
		name       string
		userID     model.UserID
		body       string
		setupStore func(*mockWorkspaceStore, *mockRuleStore)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success returns 201",
			userID: 1,
			body:   `{"match_pattern":"grocery","target_category_id":1,"priority":10}`,
			setupStore: func(ws *mockWorkspaceStore, rs *mockRuleStore) {
				ws.getMemberFunc = func(ctx context.Context, id model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: id,
						UserID:      userID,
						Role:        model.RoleEditor,
					}, nil
				}
				rs.createFunc = func(ctx context.Context, r model.CategorizationRule) (model.CategorizationRule, error) {
					return model.CategorizationRule{
						ID:               1,
						Scope:            model.ScopeWorkspace,
						WorkspaceID:      &wsID,
						MatchPattern:     r.MatchPattern,
						TargetCategoryID: r.TargetCategoryID,
						Priority:         r.Priority,
						Enabled:          true,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}, nil
				}
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			body:       `{"match_pattern":"grocery","target_category_id":1,"priority":10}`,
			setupStore: func(ws *mockWorkspaceStore, rs *mockRuleStore) {},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			rs := &mockRuleStore{}
			tt.setupStore(ws, rs)

			svc := service.NewRuleService(rs, ws)
			h := NewRuleHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/1/rules", bytes.NewBufferString(tt.body))
			} else {
				r = newAuthenticatedRequest(t, http.MethodPost, "/api/v1/workspaces/1/rules", bytes.NewBufferString(tt.body), tt.userID)
			}
			r = withChiParam(r, "id", "1")

			w := httptest.NewRecorder()
			h.CreateRule(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}

			if !tt.wantErr {
				var resp map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if resp["match_pattern"] != "grocery" {
					t.Errorf("got match_pattern %v, want 'grocery'", resp["match_pattern"])
				}
			}
		})
	}
}

func TestRuleHandler_ListRules(t *testing.T) {
	wsID := model.WorkspaceID(1)

	tests := []struct {
		name       string
		userID     model.UserID
		setupStore func(*mockWorkspaceStore, *mockRuleStore)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success returns 200",
			userID: 1,
			setupStore: func(ws *mockWorkspaceStore, rs *mockRuleStore) {
				ws.getMemberFunc = func(ctx context.Context, id model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: id,
						UserID:      userID,
						Role:        model.RoleViewer,
					}, nil
				}
				rs.listByWorkspaceFunc = func(ctx context.Context, id model.WorkspaceID) ([]model.CategorizationRule, error) {
					return []model.CategorizationRule{
						{
							ID:               1,
							Scope:            model.ScopeWorkspace,
							WorkspaceID:      &wsID,
							MatchPattern:     "grocery",
							TargetCategoryID: 1,
							Priority:         10,
							Enabled:          true,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			setupStore: func(ws *mockWorkspaceStore, rs *mockRuleStore) {},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			rs := &mockRuleStore{}
			tt.setupStore(ws, rs)

			svc := service.NewRuleService(rs, ws)
			h := NewRuleHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/1/rules", nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodGet, "/api/v1/workspaces/1/rules", nil, tt.userID)
			}
			r = withChiParam(r, "id", "1")

			w := httptest.NewRecorder()
			h.ListRules(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}

			if !tt.wantErr {
				var resp []map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if len(resp) != 1 {
					t.Errorf("got %d rules, want 1", len(resp))
				}
			}
		})
	}
}

func TestRuleHandler_DeleteRule(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		setupStore func(*mockWorkspaceStore, *mockRuleStore)
		wantStatus int
	}{
		{
			name:   "success returns 204",
			userID: 1,
			setupStore: func(ws *mockWorkspaceStore, rs *mockRuleStore) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleEditor,
					}, nil
				}
				rs.deleteFunc = func(ctx context.Context, id model.RuleID) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			setupStore: func(ws *mockWorkspaceStore, rs *mockRuleStore) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			rs := &mockRuleStore{}
			tt.setupStore(ws, rs)

			svc := service.NewRuleService(rs, ws)
			h := NewRuleHandler(svc, NewSSEHub())

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/1/rules/1", nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodDelete, "/api/v1/workspaces/1/rules/1", nil, tt.userID)
			}
			r = withChiParams(r, map[string]string{"id": "1", "ruleID": "1"})

			w := httptest.NewRecorder()
			h.DeleteRule(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}
		})
	}
}

// ── SummaryHandler tests ──

func TestSummaryHandler_GetSummary(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		query      string
		setupStore func(*mockWorkspaceStore, *mockSummaryStore, *mockFundingStoreForSummary)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success returns 200",
			userID: 1,
			query:  "?from=2025-01-01&to=2025-01-31",
			setupStore: func(ws *mockWorkspaceStore, ss *mockSummaryStore, fss *mockFundingStoreForSummary) {
				ws.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{
						WorkspaceID: wsID,
						UserID:      userID,
						Role:        model.RoleViewer,
					}, nil
				}
				ss.totalSpentFunc = func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (model.Money, error) {
					return model.NewMoney(decimal.NewFromInt(500), "PLN"), nil
				}
				ss.spentByCategoryFunc = func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.CategorySpending, error) {
					return []model.CategorySpending{
						{
							CategoryID: 1,
							Name:       "Food",
							Icon:       "utensils",
							Spent:      model.NewMoney(decimal.NewFromInt(300), "PLN"),
						},
					}, nil
				}
				ss.spentByParticipantFunc = func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.ParticipantSpending, error) {
					return []model.ParticipantSpending{
						{
							UserID:      1,
							DisplayName: "User",
							Spent:       model.NewMoney(decimal.NewFromInt(500), "PLN"),
							Funded:      model.Zero("PLN"),
							Balance:     model.Zero("PLN"),
						},
					}, nil
				}
				fss.listByWorkspaceAndUsersFunc = func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				}
				fss.getOverallBudgetsFunc = func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				}
				fss.getCategoryBudgetsFunc = func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "unauthorized returns 401",
			userID: 0,
			query:  "?from=2025-01-01&to=2025-01-31",
			setupStore: func(ws *mockWorkspaceStore, ss *mockSummaryStore, fss *mockFundingStoreForSummary) {
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:   "missing date params returns 400",
			userID: 1,
			query:  "",
			setupStore: func(ws *mockWorkspaceStore, ss *mockSummaryStore, fss *mockFundingStoreForSummary) {
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{}
			ss := &mockSummaryStore{}
			fss := &mockFundingStoreForSummary{}
			tt.setupStore(ws, ss, fss)

			svc := service.NewSummaryService(ss, fss, ws)
			h := NewSummaryHandler(svc)

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/1/summary"+tt.query, nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodGet, "/api/v1/workspaces/1/summary"+tt.query, nil, tt.userID)
			}
			r = withChiParam(r, "id", "1")

			w := httptest.NewRecorder()
			h.GetSummary(w, r)

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("got status %d, want %d", got, tt.wantStatus)
			}

			if !tt.wantErr {
				var resp map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if resp["workspace_id"] == nil {
					t.Error("got nil workspace_id, want non-nil")
				}
				totalSpent, ok := resp["total_spent"].(map[string]any)
				if !ok {
					t.Fatal("got nil total_spent, want non-nil map")
				}
				if totalSpent["amount"] != "500.00" {
					t.Errorf("got total_spent amount %v, want '500.00'", totalSpent["amount"])
				}
			}
		})
	}
}
