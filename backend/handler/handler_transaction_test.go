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
	"backend/store"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

// withChiParams adds multiple chi URL parameters to the request context.
func withChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ── Mock transaction store ──

type mockTransactionStore struct {
	createFunc                   func(ctx context.Context, t model.Transaction) (model.Transaction, error)
	getByIDFunc                  func(ctx context.Context, id model.TransactionID) (model.Transaction, error)
	listByWorkspaceFunc          func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time, limit, offset int32, filter store.ListByWorkspaceFilter) ([]model.Transaction, error)
	updateFunc                   func(ctx context.Context, id model.TransactionID, t model.Transaction) (model.Transaction, error)
	deleteFunc                   func(ctx context.Context, id model.TransactionID) error
	existsByFingerprintFunc      func(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error)
	listAllEntriesFunc           func(ctx context.Context, wsID model.WorkspaceID) ([]store.TransactionEntryRow, error)
	updateEntryCategoryFunc      func(ctx context.Context, entryID model.EntryID, catID model.CategoryID) error
	getByIDsFunc                 func(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) ([]model.Transaction, error)
	bulkCategorizeFirstEntryFunc func(ctx context.Context, txIDs []model.TransactionID, catID model.CategoryID) (int64, error)
	bulkDeleteFunc               func(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) (int64, error)
}

func (m *mockTransactionStore) Create(ctx context.Context, t model.Transaction) (model.Transaction, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, t)
	}
	return model.Transaction{}, nil
}

func (m *mockTransactionStore) GetByID(ctx context.Context, id model.TransactionID) (model.Transaction, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return model.Transaction{}, nil
}

func (m *mockTransactionStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID, from, to time.Time, limit, offset int32, filter store.ListByWorkspaceFilter) ([]model.Transaction, error) {
	if m.listByWorkspaceFunc != nil {
		return m.listByWorkspaceFunc(ctx, wsID, from, to, limit, offset, filter)
	}
	return nil, nil
}

func (m *mockTransactionStore) Update(ctx context.Context, id model.TransactionID, t model.Transaction) (model.Transaction, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, t)
	}
	return model.Transaction{}, nil
}

func (m *mockTransactionStore) Delete(ctx context.Context, id model.TransactionID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockTransactionStore) ExistsByFingerprint(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error) {
	if m.existsByFingerprintFunc != nil {
		return m.existsByFingerprintFunc(ctx, wsID, fingerprint)
	}
	return false, nil
}

func (m *mockTransactionStore) ListAllEntries(ctx context.Context, wsID model.WorkspaceID) ([]store.TransactionEntryRow, error) {
	if m.listAllEntriesFunc != nil {
		return m.listAllEntriesFunc(ctx, wsID)
	}
	return nil, nil
}

func (m *mockTransactionStore) UpdateEntryCategory(ctx context.Context, entryID model.EntryID, catID model.CategoryID) error {
	if m.updateEntryCategoryFunc != nil {
		return m.updateEntryCategoryFunc(ctx, entryID, catID)
	}
	return nil
}

func (m *mockTransactionStore) GetByIDs(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) ([]model.Transaction, error) {
	if m.getByIDsFunc != nil {
		return m.getByIDsFunc(ctx, wsID, ids)
	}
	return nil, nil
}

func (m *mockTransactionStore) BulkCategorizeFirstEntry(ctx context.Context, txIDs []model.TransactionID, catID model.CategoryID) (int64, error) {
	if m.bulkCategorizeFirstEntryFunc != nil {
		return m.bulkCategorizeFirstEntryFunc(ctx, txIDs, catID)
	}
	return 0, nil
}

func (m *mockTransactionStore) BulkDelete(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) (int64, error) {
	if m.bulkDeleteFunc != nil {
		return m.bulkDeleteFunc(ctx, wsID, ids)
	}
	return 0, nil
}

// ── Mock rule store ──

type mockRuleStore struct {
	getByIDFunc               func(ctx context.Context, id model.RuleID) (model.CategorizationRule, error)
	createFunc                func(ctx context.Context, r model.CategorizationRule) (model.CategorizationRule, error)
	updateFunc                func(ctx context.Context, id model.RuleID, pattern string, targetCatID model.CategoryID, priority int, amountMin, amountMax *decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (model.CategorizationRule, error)
	deleteFunc                func(ctx context.Context, id model.RuleID) error
	listByWorkspaceFunc       func(ctx context.Context, wsID model.WorkspaceID) ([]model.CategorizationRule, error)
	toggleEnabledFunc         func(ctx context.Context, id model.RuleID, enabled bool) (model.CategorizationRule, error)
	resolveForTransactionFunc func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, description string, amount decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (*model.CategoryID, error)
}

func (m *mockRuleStore) GetByID(ctx context.Context, id model.RuleID) (model.CategorizationRule, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	wsID := model.WorkspaceID(1)
	return model.CategorizationRule{ID: id, WorkspaceID: &wsID}, nil
}

func (m *mockRuleStore) Create(ctx context.Context, r model.CategorizationRule) (model.CategorizationRule, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, r)
	}
	return model.CategorizationRule{}, nil
}

func (m *mockRuleStore) Update(ctx context.Context, id model.RuleID, pattern string, targetCatID model.CategoryID, priority int, amountMin, amountMax *decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (model.CategorizationRule, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, pattern, targetCatID, priority, amountMin, amountMax, bankAccountID, counterpartyIBAN)
	}
	return model.CategorizationRule{}, nil
}

func (m *mockRuleStore) Delete(ctx context.Context, id model.RuleID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockRuleStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]model.CategorizationRule, error) {
	if m.listByWorkspaceFunc != nil {
		return m.listByWorkspaceFunc(ctx, wsID)
	}
	return nil, nil
}

func (m *mockRuleStore) ToggleEnabled(ctx context.Context, id model.RuleID, enabled bool) (model.CategorizationRule, error) {
	if m.toggleEnabledFunc != nil {
		return m.toggleEnabledFunc(ctx, id, enabled)
	}
	return model.CategorizationRule{}, nil
}

func (m *mockRuleStore) ResolveForTransaction(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, description string, amount decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (*model.CategoryID, error) {
	if m.resolveForTransactionFunc != nil {
		return m.resolveForTransactionFunc(ctx, wsID, userID, description, amount, bankAccountID, counterpartyIBAN)
	}
	return nil, nil
}

// ── Helper to build a transaction handler from mocks ──

type txTestMocks struct {
	txStore   *mockTransactionStore
	wsStore   *mockWorkspaceStore
	catStore  *mockCategoryStore
	ruleStore *mockRuleStore
}

func newTxTestMocks() txTestMocks {
	return txTestMocks{
		txStore:   &mockTransactionStore{},
		wsStore:   &mockWorkspaceStore{},
		catStore:  &mockCategoryStore{},
		ruleStore: &mockRuleStore{},
	}
}

func (m txTestMocks) handler() *TransactionHandler {
	ruleSvc := service.NewRuleService(m.ruleStore, m.wsStore)
	txSvc := service.NewTransactionService(m.txStore, m.wsStore, m.catStore, ruleSvc)
	return NewTransactionHandler(txSvc, NewSSEHub())
}

// ── Tests ──

func TestTransactionHandler_CreateTransaction(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name       string
		userID     model.UserID
		wsID       string
		body       string
		setup      func(*txTestMocks)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success returns 201",
			userID: 1,
			wsID:   "10",
			body: `{
				"description": "Grocery shopping",
				"date": "2025-03-15",
				"total_amount": {"amount": "100.00", "currency": "PLN"},
				"type": "expense",
				"entries": [
					{"amount": {"amount": "100.00", "currency": "PLN"}, "note": "food"}
				]
			}`,
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleEditor}, nil
				}
				m.catStore.getByNameFunc = func(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
					return model.Category{ID: 99, WorkspaceID: wsID, Name: "Uncategorized"}, nil
				}
				m.txStore.createFunc = func(ctx context.Context, tx model.Transaction) (model.Transaction, error) {
					return model.Transaction{
						ID:          1,
						WorkspaceID: tx.WorkspaceID,
						CreatedBy:   tx.CreatedBy,
						TotalAmount: tx.TotalAmount,
						Description: tx.Description,
						Date:        tx.Date,
						Source:      model.SourceManual,
						Type:        model.TypeExpense,
						Entries:     tx.Entries,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil
				}
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			wsID:       "10",
			body:       `{"description":"Test","date":"2025-03-15","total_amount":{"amount":"50","currency":"PLN"},"type":"expense","entries":[{"amount":{"amount":"50","currency":"PLN"}}]}`,
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "invalid JSON body returns 400",
			userID:     1,
			wsID:       "10",
			body:       `{invalid json}`,
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "invalid date in body returns 400",
			userID:     1,
			wsID:       "10",
			body:       `{"description":"Test","date":"not-a-date","total_amount":{"amount":"50","currency":"PLN"},"type":"expense","entries":[{"amount":{"amount":"50","currency":"PLN"}}]}`,
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:   "not a member returns 403",
			userID: 1,
			wsID:   "10",
			body:   `{"description":"Test","date":"2025-03-15","total_amount":{"amount":"50","currency":"PLN"},"type":"expense","entries":[{"amount":{"amount":"50","currency":"PLN"}}]}`,
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
				}
			},
			wantStatus: http.StatusForbidden,
			wantErr:    true,
		},
		{
			name:   "entry sum mismatch returns 400",
			userID: 1,
			wsID:   "10",
			body:   `{"description":"Test","date":"2025-03-15","total_amount":{"amount":"100.00","currency":"PLN"},"type":"expense","entries":[{"amount":{"amount":"50.00","currency":"PLN"}}]}`,
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleEditor}, nil
				}
				m.catStore.getByNameFunc = func(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
					return model.Category{ID: 99, WorkspaceID: wsID, Name: "Uncategorized"}, nil
				}
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTxTestMocks()
			tt.setup(&m)
			h := m.handler()

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+tt.wsID+"/transactions", bytes.NewBufferString(tt.body))
			} else {
				r = newAuthenticatedRequest(t, http.MethodPost, "/api/v1/workspaces/"+tt.wsID+"/transactions", bytes.NewBufferString(tt.body), tt.userID)
			}
			r = withChiParam(r, "id", tt.wsID)

			w := httptest.NewRecorder()
			h.CreateTransaction(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d (body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}

			if !tt.wantErr {
				var resp map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if resp["description"] != "Grocery shopping" {
					t.Errorf("want description 'Grocery shopping', got %v", resp["description"])
				}
			}
		})
	}
}

func TestTransactionHandler_GetTransaction(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name       string
		userID     model.UserID
		wsID       string
		txID       string
		setup      func(*txTestMocks)
		wantStatus int
	}{
		{
			name:   "success returns 200",
			userID: 1,
			wsID:   "10",
			txID:   "5",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleViewer}, nil
				}
				m.txStore.getByIDFunc = func(ctx context.Context, id model.TransactionID) (model.Transaction, error) {
					return model.Transaction{
						ID:          id,
						WorkspaceID: 10,
						CreatedBy:   1,
						TotalAmount: model.NewMoney(decimal.NewFromInt(100), "PLN"),
						Description: "Test transaction",
						Date:        now,
						Source:      model.SourceManual,
						Type:        model.TypeExpense,
						Entries: []model.Entry{
							{
								ID:         1,
								CategoryID: 99,
								Amount:     model.NewMoney(decimal.NewFromInt(100), "PLN"),
							},
						},
						CreatedAt: now,
						UpdatedAt: now,
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			wsID:       "10",
			txID:       "5",
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "not a member returns 403",
			userID: 1,
			wsID:   "10",
			txID:   "5",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
				}
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "transaction not found returns 404",
			userID: 1,
			wsID:   "10",
			txID:   "999",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleViewer}, nil
				}
				m.txStore.getByIDFunc = func(ctx context.Context, id model.TransactionID) (model.Transaction, error) {
					return model.Transaction{}, model.ErrTransactionNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTxTestMocks()
			tt.setup(&m)
			h := m.handler()

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+tt.wsID+"/transactions/"+tt.txID, nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodGet, "/api/v1/workspaces/"+tt.wsID+"/transactions/"+tt.txID, nil, tt.userID)
			}
			r = withChiParams(r, map[string]string{"id": tt.wsID, "txID": tt.txID})

			w := httptest.NewRecorder()
			h.GetTransaction(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d (body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestTransactionHandler_DeleteTransaction(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		wsID       string
		txID       string
		setup      func(*txTestMocks)
		wantStatus int
	}{
		{
			name:   "success returns 204",
			userID: 1,
			wsID:   "10",
			txID:   "5",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleEditor}, nil
				}
				m.txStore.getByIDFunc = func(ctx context.Context, id model.TransactionID) (model.Transaction, error) {
					return model.Transaction{ID: id, WorkspaceID: 10}, nil
				}
				m.txStore.deleteFunc = func(ctx context.Context, id model.TransactionID) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			wsID:       "10",
			txID:       "5",
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "not a member returns 403",
			userID: 1,
			wsID:   "10",
			txID:   "5",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
				}
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "viewer cannot delete returns 403",
			userID: 1,
			wsID:   "10",
			txID:   "5",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleViewer}, nil
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTxTestMocks()
			tt.setup(&m)
			h := m.handler()

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+tt.wsID+"/transactions/"+tt.txID, nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodDelete, "/api/v1/workspaces/"+tt.wsID+"/transactions/"+tt.txID, nil, tt.userID)
			}
			r = withChiParams(r, map[string]string{"id": tt.wsID, "txID": tt.txID})

			w := httptest.NewRecorder()
			h.DeleteTransaction(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d (body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestTransactionHandler_ListTransactions(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name       string
		userID     model.UserID
		wsID       string
		query      string
		setup      func(*txTestMocks)
		wantStatus int
		wantCount  int
	}{
		{
			name:   "success returns 200 with transactions",
			userID: 1,
			wsID:   "10",
			query:  "?from=2025-01-01&to=2025-12-31",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleViewer}, nil
				}
				m.txStore.listByWorkspaceFunc = func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time, limit, offset int32, filter store.ListByWorkspaceFilter) ([]model.Transaction, error) {
					return []model.Transaction{
						{
							ID:          1,
							WorkspaceID: wsID,
							CreatedBy:   1,
							TotalAmount: model.NewMoney(decimal.NewFromInt(50), "PLN"),
							Description: "Lunch",
							Date:        now,
							Source:      model.SourceManual,
							Type:        model.TypeExpense,
							Entries:     []model.Entry{{ID: 1, CategoryID: 99, Amount: model.NewMoney(decimal.NewFromInt(50), "PLN")}},
							CreatedAt:   now,
							UpdatedAt:   now,
						},
						{
							ID:          2,
							WorkspaceID: wsID,
							CreatedBy:   1,
							TotalAmount: model.NewMoney(decimal.NewFromInt(200), "PLN"),
							Description: "Groceries",
							Date:        now,
							Source:      model.SourceManual,
							Type:        model.TypeExpense,
							Entries:     []model.Entry{{ID: 2, CategoryID: 99, Amount: model.NewMoney(decimal.NewFromInt(200), "PLN")}},
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			wsID:       "10",
			query:      "?from=2025-01-01&to=2025-12-31",
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusUnauthorized,
			wantCount:  -1,
		},
		{
			name:       "missing from param returns 400",
			userID:     1,
			wsID:       "10",
			query:      "?to=2025-12-31",
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusBadRequest,
			wantCount:  -1,
		},
		{
			name:       "missing to param returns 400",
			userID:     1,
			wsID:       "10",
			query:      "?from=2025-01-01",
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusBadRequest,
			wantCount:  -1,
		},
		{
			name:       "invalid from date returns 400",
			userID:     1,
			wsID:       "10",
			query:      "?from=not-a-date&to=2025-12-31",
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusBadRequest,
			wantCount:  -1,
		},
		{
			name:   "empty list returns 200 with empty array",
			userID: 1,
			wsID:   "10",
			query:  "?from=2025-01-01&to=2025-12-31",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleViewer}, nil
				}
				m.txStore.listByWorkspaceFunc = func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time, limit, offset int32, filter store.ListByWorkspaceFilter) ([]model.Transaction, error) {
					return []model.Transaction{}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTxTestMocks()
			tt.setup(&m)
			h := m.handler()

			url := "/api/v1/workspaces/" + tt.wsID + "/transactions" + tt.query
			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodGet, url, nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodGet, url, nil, tt.userID)
			}
			r = withChiParam(r, "id", tt.wsID)

			w := httptest.NewRecorder()
			h.ListTransactions(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d (body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}

			if tt.wantCount >= 0 {
				var resp []map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if len(resp) != tt.wantCount {
					t.Errorf("want %d transactions, got %d", tt.wantCount, len(resp))
				}
			}
		})
	}
}

func TestTransactionHandler_ApplyRules(t *testing.T) {
	tests := []struct {
		name       string
		userID     model.UserID
		wsID       string
		setup      func(*txTestMocks)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success returns 200 with result",
			userID: 1,
			wsID:   "10",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleEditor}, nil
				}
				m.txStore.listAllEntriesFunc = func(ctx context.Context, wsID model.WorkspaceID) ([]store.TransactionEntryRow, error) {
					return []store.TransactionEntryRow{
						{
							TransactionID: 1,
							Description:   "Biedronka zakupy",
							CreatedBy:     1,
							EntryID:       10,
							CategoryID:    99,
							Type:          model.TypeExpense,
							TotalAmount:   decimal.NewFromInt(100),
						},
					}, nil
				}
				m.catStore.getByNameFunc = func(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
					return model.Category{ID: 99, WorkspaceID: wsID, Name: "Uncategorized"}, nil
				}
				m.ruleStore.resolveForTransactionFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, description string, amount decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (*model.CategoryID, error) {
					catID := model.CategoryID(5)
					return &catID, nil
				}
				m.txStore.updateEntryCategoryFunc = func(ctx context.Context, entryID model.EntryID, catID model.CategoryID) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "unauthorized returns 401",
			userID:     0,
			wsID:       "10",
			setup:      func(m *txTestMocks) {},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:   "not a member returns 403",
			userID: 1,
			wsID:   "10",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
				}
			},
			wantStatus: http.StatusForbidden,
			wantErr:    true,
		},
		{
			name:   "viewer cannot apply rules returns 403",
			userID: 1,
			wsID:   "10",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleViewer}, nil
				}
			},
			wantStatus: http.StatusForbidden,
			wantErr:    true,
		},
		{
			name:   "no entries returns 200 with zero counts",
			userID: 1,
			wsID:   "10",
			setup: func(m *txTestMocks) {
				m.wsStore.getMemberFunc = func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID) (model.WorkspaceMember, error) {
					return model.WorkspaceMember{WorkspaceID: wsID, UserID: userID, Role: model.RoleEditor}, nil
				}
				m.txStore.listAllEntriesFunc = func(ctx context.Context, wsID model.WorkspaceID) ([]store.TransactionEntryRow, error) {
					return []store.TransactionEntryRow{}, nil
				}
				m.catStore.getByNameFunc = func(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
					return model.Category{ID: 99, WorkspaceID: wsID, Name: "Uncategorized"}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTxTestMocks()
			tt.setup(&m)
			h := m.handler()

			var r *http.Request
			if tt.userID == 0 {
				r = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+tt.wsID+"/transactions/apply-rules", nil)
			} else {
				r = newAuthenticatedRequest(t, http.MethodPost, "/api/v1/workspaces/"+tt.wsID+"/transactions/apply-rules", nil, tt.userID)
			}
			r = withChiParam(r, "id", tt.wsID)

			w := httptest.NewRecorder()
			h.ApplyRules(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d (body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}

			if !tt.wantErr {
				var resp map[string]any
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decoding response: %v", err)
				}
				if _, ok := resp["total"]; !ok {
					t.Error("want 'total' field in response")
				}
				if _, ok := resp["updated"]; !ok {
					t.Error("want 'updated' field in response")
				}
			}
		})
	}
}
