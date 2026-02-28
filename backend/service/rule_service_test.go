package service

import (
	"context"
	"errors"
	"testing"

	"backend/model"

	"github.com/shopspring/decimal"
)

// ── Mock rule store ──

type mockRuleStore struct {
	GetByIDFunc               func(ctx context.Context, id model.RuleID) (model.CategorizationRule, error)
	CreateFunc                func(ctx context.Context, r model.CategorizationRule) (model.CategorizationRule, error)
	UpdateFunc                func(ctx context.Context, id model.RuleID, pattern string, targetCatID model.CategoryID, priority int, amountMin, amountMax *decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (model.CategorizationRule, error)
	DeleteFunc                func(ctx context.Context, id model.RuleID) error
	ListByWorkspaceFunc       func(ctx context.Context, wsID model.WorkspaceID) ([]model.CategorizationRule, error)
	ToggleEnabledFunc         func(ctx context.Context, id model.RuleID, enabled bool) (model.CategorizationRule, error)
	ResolveForTransactionFunc func(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, description string, amount decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (*model.CategoryID, error)
}

func (m *mockRuleStore) GetByID(ctx context.Context, id model.RuleID) (model.CategorizationRule, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	wsID := model.WorkspaceID(1)
	return model.CategorizationRule{ID: id, WorkspaceID: &wsID}, nil
}

func (m *mockRuleStore) Create(ctx context.Context, r model.CategorizationRule) (model.CategorizationRule, error) {
	return m.CreateFunc(ctx, r)
}

func (m *mockRuleStore) Update(ctx context.Context, id model.RuleID, pattern string, targetCatID model.CategoryID, priority int, amountMin, amountMax *decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (model.CategorizationRule, error) {
	return m.UpdateFunc(ctx, id, pattern, targetCatID, priority, amountMin, amountMax, bankAccountID, counterpartyIBAN)
}

func (m *mockRuleStore) Delete(ctx context.Context, id model.RuleID) error {
	return m.DeleteFunc(ctx, id)
}

func (m *mockRuleStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]model.CategorizationRule, error) {
	return m.ListByWorkspaceFunc(ctx, wsID)
}

func (m *mockRuleStore) ToggleEnabled(ctx context.Context, id model.RuleID, enabled bool) (model.CategorizationRule, error) {
	return m.ToggleEnabledFunc(ctx, id, enabled)
}

func (m *mockRuleStore) ResolveForTransaction(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, description string, amount decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (*model.CategoryID, error) {
	return m.ResolveForTransactionFunc(ctx, wsID, userID, description, amount, bankAccountID, counterpartyIBAN)
}

// ── Tests ──

func TestRuleService_Create(t *testing.T) {
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
			rs := &mockRuleStore{
				CreateFunc: func(_ context.Context, r model.CategorizationRule) (model.CategorizationRule, error) {
					r.ID = 1
					if r.Scope != model.ScopeWorkspace {
						t.Errorf("got scope %v, want %v", r.Scope, model.ScopeWorkspace)
					}
					if !r.Enabled {
						t.Errorf("got Enabled %v, want true", r.Enabled)
					}
					return r, nil
				},
			}
			svc := NewRuleService(rs, ws)

			input := CreateRuleInput{
				MatchPattern:     "grocery",
				TargetCategoryID: 10,
				Priority:         100,
			}
			got, err := svc.Create(context.Background(), 1, 1, input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.MatchPattern != input.MatchPattern {
				t.Errorf("got pattern %q, want %q", got.MatchPattern, input.MatchPattern)
			}
		})
	}
}

func TestRuleService_Update(t *testing.T) {
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
			rs := &mockRuleStore{
				UpdateFunc: func(_ context.Context, id model.RuleID, pattern string, targetCatID model.CategoryID, priority int, amountMin, amountMax *decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (model.CategorizationRule, error) {
					return model.CategorizationRule{
						ID:               id,
						MatchPattern:     pattern,
						TargetCategoryID: targetCatID,
						Priority:         priority,
						AmountMin:        amountMin,
						AmountMax:        amountMax,
						BankAccountID:    bankAccountID,
					}, nil
				},
			}
			svc := NewRuleService(rs, ws)

			input := UpdateRuleInput{
				MatchPattern:     "updated pattern",
				TargetCategoryID: 20,
				Priority:         200,
			}
			_, err := svc.Update(context.Background(), 1, 1, 5, input)
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

func TestRuleService_Delete(t *testing.T) {
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
			rs := &mockRuleStore{
				DeleteFunc: func(_ context.Context, _ model.RuleID) error {
					return nil
				},
			}
			svc := NewRuleService(rs, ws)

			err := svc.Delete(context.Background(), 1, 1, 5)
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

func TestRuleService_Toggle(t *testing.T) {
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
			rs := &mockRuleStore{
				ToggleEnabledFunc: func(_ context.Context, id model.RuleID, enabled bool) (model.CategorizationRule, error) {
					return model.CategorizationRule{
						ID:      id,
						Enabled: enabled,
					}, nil
				},
			}
			svc := NewRuleService(rs, ws)

			_, err := svc.Toggle(context.Background(), 1, 1, 5, false)
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

func TestRuleService_ListByWorkspace(t *testing.T) {
	tests := map[string]struct {
		role    model.MemberRole
		wantErr error
	}{
		"viewer allowed": {
			role: model.RoleViewer,
		},
		"non-member rejected": {
			role:    model.MemberRole(""), // invalid role with level 0
			wantErr: model.ErrNotWorkspaceMember,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					if tc.role == "" {
						return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
					}
					return memberWithRole(tc.role), nil
				},
			}
			rs := &mockRuleStore{
				ListByWorkspaceFunc: func(_ context.Context, _ model.WorkspaceID) ([]model.CategorizationRule, error) {
					return []model.CategorizationRule{
						{ID: 1, MatchPattern: "pattern1"},
						{ID: 2, MatchPattern: "pattern2"},
					}, nil
				},
			}
			svc := NewRuleService(rs, ws)

			_, err := svc.ListByWorkspace(context.Background(), 1, 1)
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

func TestRuleService_ResolveCategory(t *testing.T) {
	catID := model.CategoryID(10)
	rs := &mockRuleStore{
		ResolveForTransactionFunc: func(_ context.Context, wsID model.WorkspaceID, userID model.UserID, description string, amount decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (*model.CategoryID, error) {
			if description == "grocery store" {
				return &catID, nil
			}
			return nil, nil
		},
	}
	svc := NewRuleService(rs, &mockWorkspaceStore{})

	got, err := svc.ResolveCategory(context.Background(), 1, 1, "grocery store", decimal.NewFromInt(100), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil, want category ID")
	}
	if *got != catID {
		t.Errorf("got %v, want %v", *got, catID)
	}

	// Test no match
	got, err = svc.ResolveCategory(context.Background(), 1, 1, "unknown", decimal.NewFromInt(100), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("got %v, want nil", *got)
	}
}

func TestRuleService_ValidateAmountRange(t *testing.T) {
	tests := map[string]struct {
		min     *decimal.Decimal
		max     *decimal.Decimal
		wantErr bool
	}{
		"both nil": {
			min: nil, max: nil, wantErr: false,
		},
		"valid range": {
			min:     func() *decimal.Decimal { d := decimal.NewFromInt(100); return &d }(),
			max:     func() *decimal.Decimal { d := decimal.NewFromInt(200); return &d }(),
			wantErr: false,
		},
		"min equals max": {
			min:     func() *decimal.Decimal { d := decimal.NewFromInt(100); return &d }(),
			max:     func() *decimal.Decimal { d := decimal.NewFromInt(100); return &d }(),
			wantErr: false,
		},
		"min greater than max": {
			min:     func() *decimal.Decimal { d := decimal.NewFromInt(200); return &d }(),
			max:     func() *decimal.Decimal { d := decimal.NewFromInt(100); return &d }(),
			wantErr: true,
		},
		"negative min": {
			min:     func() *decimal.Decimal { d := decimal.NewFromInt(-10); return &d }(),
			max:     nil,
			wantErr: true,
		},
		"negative max": {
			min:     nil,
			max:     func() *decimal.Decimal { d := decimal.NewFromInt(-10); return &d }(),
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(model.RoleEditor), nil
				},
			}
			rs := &mockRuleStore{
				CreateFunc: func(_ context.Context, r model.CategorizationRule) (model.CategorizationRule, error) {
					r.ID = 1
					return r, nil
				},
			}
			svc := NewRuleService(rs, ws)

			_, err := svc.Create(context.Background(), 1, 1, CreateRuleInput{
				MatchPattern:     "test",
				TargetCategoryID: 10,
				Priority:         100,
				AmountMin:        tc.min,
				AmountMax:        tc.max,
			})
			if tc.wantErr && err == nil {
				t.Error("want error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("want no error, got %v", err)
			}
		})
	}
}
