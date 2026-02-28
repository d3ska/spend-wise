package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/model"
	"backend/store"

	"github.com/shopspring/decimal"
)

// ── Mock stores ──

type mockTransactionStore struct {
	CreateFunc                   func(ctx context.Context, t model.Transaction) (model.Transaction, error)
	GetByIDFunc                  func(ctx context.Context, id model.TransactionID) (model.Transaction, error)
	ListByWorkspaceFunc          func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time, limit, offset int32, filter store.ListByWorkspaceFilter) ([]model.Transaction, error)
	UpdateFunc                   func(ctx context.Context, id model.TransactionID, t model.Transaction) (model.Transaction, error)
	DeleteFunc                   func(ctx context.Context, id model.TransactionID) error
	ExistsByFingerprintFunc      func(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error)
	ListAllEntriesFunc           func(ctx context.Context, wsID model.WorkspaceID) ([]store.TransactionEntryRow, error)
	UpdateEntryCategoryFunc      func(ctx context.Context, entryID model.EntryID, catID model.CategoryID) error
	GetByIDsFunc                 func(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) ([]model.Transaction, error)
	BulkCategorizeFirstEntryFunc func(ctx context.Context, txIDs []model.TransactionID, catID model.CategoryID) (int64, error)
	BulkDeleteFunc               func(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) (int64, error)
}

func (m *mockTransactionStore) Create(ctx context.Context, t model.Transaction) (model.Transaction, error) {
	return m.CreateFunc(ctx, t)
}

func (m *mockTransactionStore) GetByID(ctx context.Context, id model.TransactionID) (model.Transaction, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *mockTransactionStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID, from, to time.Time, limit, offset int32, filter store.ListByWorkspaceFilter) ([]model.Transaction, error) {
	return m.ListByWorkspaceFunc(ctx, wsID, from, to, limit, offset, filter)
}

func (m *mockTransactionStore) Update(ctx context.Context, id model.TransactionID, t model.Transaction) (model.Transaction, error) {
	return m.UpdateFunc(ctx, id, t)
}

func (m *mockTransactionStore) Delete(ctx context.Context, id model.TransactionID) error {
	return m.DeleteFunc(ctx, id)
}

func (m *mockTransactionStore) ExistsByFingerprint(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error) {
	return m.ExistsByFingerprintFunc(ctx, wsID, fingerprint)
}

func (m *mockTransactionStore) ListAllEntries(ctx context.Context, wsID model.WorkspaceID) ([]store.TransactionEntryRow, error) {
	return m.ListAllEntriesFunc(ctx, wsID)
}

func (m *mockTransactionStore) UpdateEntryCategory(ctx context.Context, entryID model.EntryID, catID model.CategoryID) error {
	return m.UpdateEntryCategoryFunc(ctx, entryID, catID)
}

func (m *mockTransactionStore) GetByIDs(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) ([]model.Transaction, error) {
	return m.GetByIDsFunc(ctx, wsID, ids)
}

func (m *mockTransactionStore) BulkCategorizeFirstEntry(ctx context.Context, txIDs []model.TransactionID, catID model.CategoryID) (int64, error) {
	return m.BulkCategorizeFirstEntryFunc(ctx, txIDs, catID)
}

func (m *mockTransactionStore) BulkDelete(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) (int64, error) {
	return m.BulkDeleteFunc(ctx, wsID, ids)
}

// ── Helpers ──

func mustMoney(amount int64, currency string) model.Money {
	return model.NewMoney(decimal.NewFromInt(amount), currency)
}

// ── Tests ──

func TestTransactionService_Create(t *testing.T) {
	tests := map[string]struct {
		role      model.MemberRole
		input     CreateTransactionInput
		catExists bool
		wantErr   error
	}{
		"valid entries with editor role": {
			role: model.RoleEditor,
			input: CreateTransactionInput{
				Description: "Groceries",
				Date:        time.Now(),
				TotalAmount: mustMoney(100, "PLN"),
				Type:        model.TypeExpense,
				Entries: []CreateEntryInput{
					{CategoryID: 5, Amount: mustMoney(100, "PLN")},
				},
			},
		},
		"default type to expense when empty": {
			role: model.RoleEditor,
			input: CreateTransactionInput{
				Description: "Groceries",
				Date:        time.Now(),
				TotalAmount: mustMoney(100, "PLN"),
				Type:        "",
				Entries: []CreateEntryInput{
					{CategoryID: 5, Amount: mustMoney(100, "PLN")},
				},
			},
		},
		"resolves zero CategoryID to Uncategorized": {
			role: model.RoleEditor,
			input: CreateTransactionInput{
				Description: "Groceries",
				Date:        time.Now(),
				TotalAmount: mustMoney(100, "PLN"),
				Type:        model.TypeExpense,
				Entries: []CreateEntryInput{
					{CategoryID: 0, Amount: mustMoney(100, "PLN")},
				},
			},
			catExists: true,
		},
		"mismatched entries sum": {
			role: model.RoleEditor,
			input: CreateTransactionInput{
				Description: "Groceries",
				Date:        time.Now(),
				TotalAmount: mustMoney(100, "PLN"),
				Type:        model.TypeExpense,
				Entries: []CreateEntryInput{
					{CategoryID: 5, Amount: mustMoney(50, "PLN")},
				},
			},
			wantErr: model.ErrEntrySumMismatch,
		},
		"viewer rejected": {
			role: model.RoleViewer,
			input: CreateTransactionInput{
				Description: "Groceries",
				Date:        time.Now(),
				TotalAmount: mustMoney(100, "PLN"),
				Type:        model.TypeExpense,
				Entries: []CreateEntryInput{
					{CategoryID: 5, Amount: mustMoney(100, "PLN")},
				},
			},
			wantErr: model.ErrInsufficientPermission,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			txStore := &mockTransactionStore{
				CreateFunc: func(_ context.Context, t model.Transaction) (model.Transaction, error) {
					t.ID = 1
					return t, nil
				},
			}
			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
			}
			catStore := &mockCategoryStore{
				GetBySlugFunc: func(_ context.Context, _ model.WorkspaceID, slug string) (model.Category, error) {
					if tc.catExists && slug == "uncategorized" {
						return model.Category{ID: 999, Name: "Uncategorized"}, nil
					}
					return model.Category{}, model.ErrCategoryNotFound
				},
				CreateFunc: func(_ context.Context, cat model.Category) (model.Category, error) {
					cat.ID = 999
					return cat, nil
				},
			}

			svc := NewTransactionService(txStore, wsStore, catStore, nil)

			got, err := svc.Create(context.Background(), 1, 1, tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Description != tc.input.Description {
				t.Errorf("got description %q, want %q", got.Description, tc.input.Description)
			}
			if tc.input.Type == "" && got.Type != model.TypeExpense {
				t.Errorf("got type %q, want %q", got.Type, model.TypeExpense)
			}
		})
	}
}

func TestTransactionService_GetTransaction(t *testing.T) {
	tests := map[string]struct {
		role    model.MemberRole
		wantErr error
	}{
		"viewer allowed": {
			role: model.RoleViewer,
		},
		"non-member": {
			role:    model.RoleViewer,
			wantErr: model.ErrNotWorkspaceMember,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			txStore := &mockTransactionStore{
				GetByIDFunc: func(_ context.Context, _ model.TransactionID) (model.Transaction, error) {
					return model.Transaction{ID: 1, WorkspaceID: 1, Description: "Test"}, nil
				},
			}
			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					if tc.wantErr == model.ErrNotWorkspaceMember {
						return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
					}
					return memberWithRole(tc.role), nil
				},
			}

			svc := NewTransactionService(txStore, wsStore, &mockCategoryStore{}, nil)

			_, err := svc.GetTransaction(context.Background(), 1, 1, 1)
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

func TestTransactionService_ListTransactions(t *testing.T) {
	tests := map[string]struct {
		role    model.MemberRole
		wantErr error
	}{
		"viewer allowed": {
			role: model.RoleViewer,
		},
		"non-member": {
			role:    model.RoleViewer,
			wantErr: model.ErrNotWorkspaceMember,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			txStore := &mockTransactionStore{
				ListByWorkspaceFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time, _, _ int32, _ store.ListByWorkspaceFilter) ([]model.Transaction, error) {
					return []model.Transaction{{ID: 1}}, nil
				},
			}
			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					if tc.wantErr == model.ErrNotWorkspaceMember {
						return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
					}
					return memberWithRole(tc.role), nil
				},
			}

			svc := NewTransactionService(txStore, wsStore, &mockCategoryStore{}, nil)

			_, err := svc.ListTransactions(context.Background(), 1, 1, ListTransactionsInput{
				From: time.Now().Add(-24 * time.Hour),
				To:   time.Now(),
			})
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

func TestTransactionService_DeleteTransaction(t *testing.T) {
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
			txStore := &mockTransactionStore{
				GetByIDFunc: func(_ context.Context, _ model.TransactionID) (model.Transaction, error) {
					return model.Transaction{ID: 1, WorkspaceID: 1}, nil
				},
				DeleteFunc: func(_ context.Context, _ model.TransactionID) error {
					return nil
				},
			}
			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
			}

			svc := NewTransactionService(txStore, wsStore, &mockCategoryStore{}, nil)

			err := svc.DeleteTransaction(context.Background(), 1, 1, 1)
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

func TestTransactionService_UpdateTransaction(t *testing.T) {
	tests := map[string]struct {
		role               model.MemberRole
		existingWorkspace  model.WorkspaceID
		requestedWorkspace model.WorkspaceID
		wantErr            error
	}{
		"editor allowed": {
			role:               model.RoleEditor,
			existingWorkspace:  1,
			requestedWorkspace: 1,
		},
		"viewer rejected": {
			role:               model.RoleViewer,
			existingWorkspace:  1,
			requestedWorkspace: 1,
			wantErr:            model.ErrInsufficientPermission,
		},
		"workspace mismatch": {
			role:               model.RoleEditor,
			existingWorkspace:  1,
			requestedWorkspace: 2,
			wantErr:            model.ErrTransactionNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			txStore := &mockTransactionStore{
				GetByIDFunc: func(_ context.Context, _ model.TransactionID) (model.Transaction, error) {
					return model.Transaction{
						ID:          1,
						WorkspaceID: tc.existingWorkspace,
						TotalAmount: mustMoney(100, "PLN"),
					}, nil
				},
				UpdateFunc: func(_ context.Context, _ model.TransactionID, t model.Transaction) (model.Transaction, error) {
					t.ID = 1
					return t, nil
				},
			}
			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
			}
			catStore := &mockCategoryStore{
				GetBySlugFunc: func(_ context.Context, _ model.WorkspaceID, slug string) (model.Category, error) {
					if slug == "uncategorized" {
						return model.Category{ID: 999, Name: "Uncategorized"}, nil
					}
					return model.Category{}, model.ErrCategoryNotFound
				},
				CreateFunc: func(_ context.Context, cat model.Category) (model.Category, error) {
					cat.ID = 999
					return cat, nil
				},
			}

			svc := NewTransactionService(txStore, wsStore, catStore, nil)

			_, err := svc.UpdateTransaction(context.Background(), 1, tc.requestedWorkspace, 1, UpdateTransactionInput{
				Description: "Updated",
				Date:        time.Now(),
				Type:        model.TypeExpense,
				Entries: []CreateEntryInput{
					{CategoryID: 5, Amount: mustMoney(100, "PLN")},
				},
			})
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

func TestTransactionService_ImportTransactions(t *testing.T) {
	tests := map[string]struct {
		fingerprints     map[string]bool
		ruleResolves     bool
		wantImported     int
		wantSkipped      int
		wantTransactions int
	}{
		"skips duplicate fingerprint": {
			fingerprints: map[string]bool{
				"fp1": true,
				"fp2": false,
			},
			wantImported:     1,
			wantSkipped:      1,
			wantTransactions: 1,
		},
		"creates new transactions": {
			fingerprints: map[string]bool{
				"fp1": false,
				"fp2": false,
			},
			wantImported:     2,
			wantSkipped:      0,
			wantTransactions: 2,
		},
		"auto-categorizes via rule service": {
			fingerprints: map[string]bool{
				"fp1": false,
			},
			ruleResolves:     true,
			wantImported:     1,
			wantSkipped:      0,
			wantTransactions: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var createdCount int
			txStore := &mockTransactionStore{
				ExistsByFingerprintFunc: func(_ context.Context, _ model.WorkspaceID, fp string) (bool, error) {
					return tc.fingerprints[fp], nil
				},
				CreateFunc: func(_ context.Context, t model.Transaction) (model.Transaction, error) {
					createdCount++
					t.ID = model.TransactionID(createdCount)
					return t, nil
				},
			}
			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(model.RoleEditor), nil
				},
			}

			var ruleSvc *RuleService
			if tc.ruleResolves {
				ruleStore := &mockRuleStore{
					ResolveForTransactionFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID, _ string, _ decimal.Decimal, _ *model.BankAccountID, _ *string) (*model.CategoryID, error) {
						catID := model.CategoryID(10)
						return &catID, nil
					},
				}
				ruleSvc = NewRuleService(ruleStore, wsStore)
			}

			svc := NewTransactionService(txStore, wsStore, &mockCategoryStore{}, ruleSvc)

			inputs := []ImportTransactionInput{
				{
					CreateTransactionInput: CreateTransactionInput{
						Description: "Import 1",
						Date:        time.Now(),
						TotalAmount: mustMoney(100, "PLN"),
						Entries: []CreateEntryInput{
							{CategoryID: 0, Amount: mustMoney(100, "PLN")},
						},
					},
					Fingerprint: "fp1",
				},
			}

			if len(tc.fingerprints) > 1 {
				inputs = append(inputs, ImportTransactionInput{
					CreateTransactionInput: CreateTransactionInput{
						Description: "Import 2",
						Date:        time.Now(),
						TotalAmount: mustMoney(200, "PLN"),
						Entries: []CreateEntryInput{
							{CategoryID: 0, Amount: mustMoney(200, "PLN")},
						},
					},
					Fingerprint: "fp2",
				})
			}

			result, err := svc.ImportTransactions(context.Background(), 1, 1, inputs)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Imported != tc.wantImported {
				t.Errorf("got imported %d, want %d", result.Imported, tc.wantImported)
			}
			if result.Skipped != tc.wantSkipped {
				t.Errorf("got skipped %d, want %d", result.Skipped, tc.wantSkipped)
			}
			if len(result.Transactions) != tc.wantTransactions {
				t.Errorf("got %d transactions, want %d", len(result.Transactions), tc.wantTransactions)
			}
		})
	}
}

func TestTransactionService_ApplyRules(t *testing.T) {
	tests := map[string]struct {
		entries      []store.TransactionEntryRow
		ruleMatches  map[string]*model.CategoryID
		wantUpdated  int
		wantTotal    int
		skipTransfer bool
	}{
		"re-categorizes matched entries": {
			entries: []store.TransactionEntryRow{
				{EntryID: 1, Description: "Coffee", CategoryID: 5, Type: model.TypeExpense, TotalAmount: decimal.NewFromInt(50)},
				{EntryID: 2, Description: "Groceries", CategoryID: 6, Type: model.TypeExpense, TotalAmount: decimal.NewFromInt(100)},
			},
			ruleMatches: map[string]*model.CategoryID{
				"Coffee":    func() *model.CategoryID { id := model.CategoryID(10); return &id }(),
				"Groceries": func() *model.CategoryID { id := model.CategoryID(11); return &id }(),
			},
			wantUpdated: 2,
			wantTotal:   2,
		},
		"skips income": {
			entries: []store.TransactionEntryRow{
				{EntryID: 1, Description: "Salary", CategoryID: 5, Type: model.TypeIncome, TotalAmount: decimal.NewFromInt(5000)},
				{EntryID: 2, Description: "Coffee", CategoryID: 6, Type: model.TypeExpense, TotalAmount: decimal.NewFromInt(20)},
			},
			ruleMatches: map[string]*model.CategoryID{
				"Salary": func() *model.CategoryID { id := model.CategoryID(10); return &id }(),
				"Coffee": func() *model.CategoryID { id := model.CategoryID(11); return &id }(),
			},
			wantUpdated: 1,
			wantTotal:   2,
		},
		"sets unmatched to Uncategorized": {
			entries: []store.TransactionEntryRow{
				{EntryID: 1, Description: "Random", CategoryID: 5, Type: model.TypeExpense, TotalAmount: decimal.NewFromInt(75)},
			},
			ruleMatches: map[string]*model.CategoryID{
				"Random": nil,
			},
			wantUpdated: 1,
			wantTotal:   1,
		},
		"skips transfers": {
			entries: []store.TransactionEntryRow{
				{EntryID: 1, Description: "Transfer", CategoryID: 5, Type: model.TypeTransfer, TotalAmount: decimal.NewFromInt(500)},
				{EntryID: 2, Description: "Coffee", CategoryID: 6, Type: model.TypeExpense, TotalAmount: decimal.NewFromInt(15)},
			},
			ruleMatches: map[string]*model.CategoryID{
				"Transfer": func() *model.CategoryID { id := model.CategoryID(10); return &id }(),
				"Coffee":   func() *model.CategoryID { id := model.CategoryID(11); return &id }(),
			},
			wantUpdated: 1,
			wantTotal:   2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var updatedEntries []model.EntryID
			txStore := &mockTransactionStore{
				ListAllEntriesFunc: func(_ context.Context, _ model.WorkspaceID) ([]store.TransactionEntryRow, error) {
					return tc.entries, nil
				},
				UpdateEntryCategoryFunc: func(_ context.Context, entryID model.EntryID, catID model.CategoryID) error {
					updatedEntries = append(updatedEntries, entryID)
					return nil
				},
			}
			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(model.RoleEditor), nil
				},
			}
			catStore := &mockCategoryStore{
				GetBySlugFunc: func(_ context.Context, _ model.WorkspaceID, slug string) (model.Category, error) {
					if slug == "uncategorized" {
						return model.Category{ID: 999, Name: "Uncategorized"}, nil
					}
					return model.Category{}, model.ErrCategoryNotFound
				},
				CreateFunc: func(_ context.Context, cat model.Category) (model.Category, error) {
					cat.ID = 999
					return cat, nil
				},
			}
			ruleStore := &mockRuleStore{
				ResolveForTransactionFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID, desc string, _ decimal.Decimal, _ *model.BankAccountID, _ *string) (*model.CategoryID, error) {
					return tc.ruleMatches[desc], nil
				},
			}

			ruleSvc := NewRuleService(ruleStore, wsStore)
			svc := NewTransactionService(txStore, wsStore, catStore, ruleSvc)

			result, err := svc.ApplyRules(context.Background(), 1, 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Updated != tc.wantUpdated {
				t.Errorf("got updated %d, want %d", result.Updated, tc.wantUpdated)
			}
			if result.Total != tc.wantTotal {
				t.Errorf("got total %d, want %d", result.Total, tc.wantTotal)
			}
		})
	}
}

func TestTransactionService_BulkCategorizeTransactions(t *testing.T) {
	tests := map[string]struct {
		role        model.MemberRole
		ids         []model.TransactionID
		categoryID  model.CategoryID
		catExists   bool
		catWsID     model.WorkspaceID
		matchingTxs int
		wantErr     error
		wantUpdated int64
	}{
		"editor allowed — updates all": {
			role:        model.RoleEditor,
			ids:         []model.TransactionID{1, 2, 3},
			categoryID:  5,
			catExists:   true,
			catWsID:     1,
			matchingTxs: 3,
			wantUpdated: 3,
		},
		"viewer rejected": {
			role:       model.RoleViewer,
			ids:        []model.TransactionID{1},
			categoryID: 5,
			wantErr:    model.ErrInsufficientPermission,
		},
		"empty ids returns zero": {
			role:        model.RoleEditor,
			ids:         []model.TransactionID{},
			categoryID:  5,
			wantUpdated: 0,
		},
		"invalid category not found": {
			role:       model.RoleEditor,
			ids:        []model.TransactionID{1},
			categoryID: 999,
			catExists:  false,
			wantErr:    model.ErrCategoryNotFound,
		},
		"category from wrong workspace": {
			role:       model.RoleEditor,
			ids:        []model.TransactionID{1},
			categoryID: 5,
			catExists:  true,
			catWsID:    99,
			wantErr:    model.ErrCategoryNotFound,
		},
		"IDOR — transaction not in workspace": {
			role:        model.RoleEditor,
			ids:         []model.TransactionID{1, 2, 3},
			categoryID:  5,
			catExists:   true,
			catWsID:     1,
			matchingTxs: 2, // only 2 of 3 match
			wantErr:     model.ErrTransactionNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			txStore := &mockTransactionStore{
				GetByIDsFunc: func(_ context.Context, _ model.WorkspaceID, ids []model.TransactionID) ([]model.Transaction, error) {
					result := make([]model.Transaction, 0, tc.matchingTxs)
					for i := range tc.matchingTxs {
						result = append(result, model.Transaction{ID: ids[i], WorkspaceID: 1})
					}
					return result, nil
				},
				BulkCategorizeFirstEntryFunc: func(_ context.Context, txIDs []model.TransactionID, _ model.CategoryID) (int64, error) {
					return int64(len(txIDs)), nil
				},
			}
			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
			}
			catStore := &mockCategoryStore{
				GetByIDFunc: func(_ context.Context, id model.CategoryID) (model.Category, error) {
					if tc.catExists {
						return model.Category{ID: id, WorkspaceID: tc.catWsID}, nil
					}
					return model.Category{}, model.ErrCategoryNotFound
				},
			}

			svc := NewTransactionService(txStore, wsStore, catStore, nil)

			updated, err := svc.BulkCategorizeTransactions(context.Background(), 1, 1, tc.ids, tc.categoryID)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if updated != tc.wantUpdated {
				t.Errorf("got updated %d, want %d", updated, tc.wantUpdated)
			}
		})
	}
}

func TestTransactionService_BulkDeleteTransactions(t *testing.T) {
	tests := map[string]struct {
		role        model.MemberRole
		ids         []model.TransactionID
		matchingTxs int
		wantErr     error
		wantDeleted int64
	}{
		"editor allowed — deletes all": {
			role:        model.RoleEditor,
			ids:         []model.TransactionID{1, 2, 3},
			matchingTxs: 3,
			wantDeleted: 3,
		},
		"viewer rejected": {
			role:    model.RoleViewer,
			ids:     []model.TransactionID{1},
			wantErr: model.ErrInsufficientPermission,
		},
		"empty ids returns zero": {
			role:        model.RoleEditor,
			ids:         []model.TransactionID{},
			wantDeleted: 0,
		},
		"IDOR — transaction not in workspace": {
			role:        model.RoleEditor,
			ids:         []model.TransactionID{1, 2, 3},
			matchingTxs: 2,
			wantErr:     model.ErrTransactionNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			txStore := &mockTransactionStore{
				GetByIDsFunc: func(_ context.Context, _ model.WorkspaceID, ids []model.TransactionID) ([]model.Transaction, error) {
					result := make([]model.Transaction, 0, tc.matchingTxs)
					for i := range tc.matchingTxs {
						result = append(result, model.Transaction{ID: ids[i], WorkspaceID: 1})
					}
					return result, nil
				},
				BulkDeleteFunc: func(_ context.Context, _ model.WorkspaceID, ids []model.TransactionID) (int64, error) {
					return int64(len(ids)), nil
				},
			}
			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(tc.role), nil
				},
			}

			svc := NewTransactionService(txStore, wsStore, &mockCategoryStore{}, nil)

			deleted, err := svc.BulkDeleteTransactions(context.Background(), 1, 1, tc.ids)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if deleted != tc.wantDeleted {
				t.Errorf("got deleted %d, want %d", deleted, tc.wantDeleted)
			}
		})
	}
}
