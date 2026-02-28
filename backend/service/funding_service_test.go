package service

import (
	"context"
	"errors"
	"testing"

	"backend/model"

	"github.com/shopspring/decimal"
)

// ── Mock funding store ──

type mockFundingStore struct {
	UpsertFunc          func(ctx context.Context, f model.Funding) (model.Funding, error)
	ListByWorkspaceFunc func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
	DeleteFunc          func(ctx context.Context, id model.FundingID, wsID model.WorkspaceID) error
}

func (m *mockFundingStore) Upsert(ctx context.Context, f model.Funding) (model.Funding, error) {
	return m.UpsertFunc(ctx, f)
}

func (m *mockFundingStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	return m.ListByWorkspaceFunc(ctx, wsID, fromMonth, toMonth)
}

func (m *mockFundingStore) Delete(ctx context.Context, id model.FundingID, wsID model.WorkspaceID) error {
	return m.DeleteFunc(ctx, id, wsID)
}

// ── Tests ──

func TestFundingService_Record(t *testing.T) {
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
			fs := &mockFundingStore{
				ListByWorkspaceFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				},
				UpsertFunc: func(_ context.Context, f model.Funding) (model.Funding, error) {
					f.ID = 1
					return f, nil
				},
			}
			svc := NewFundingService(fs, ws)

			input := RecordFundingInput{
				UserID:     1,
				CategoryID: nil,
				YearMonth:  "2026-01",
				Amount:     model.NewMoney(decimal.NewFromInt(5000), "PLN"),
			}
			_, err := svc.Record(context.Background(), 1, 1, input)
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

func TestFundingService_Record_BudgetValidation(t *testing.T) {
	catID1 := model.CategoryID(10)
	catID2 := model.CategoryID(20)

	tests := map[string]struct {
		existing []model.Funding
		input    RecordFundingInput
		wantErr  error
	}{
		"category within overall limit": {
			existing: []model.Funding{
				{CategoryID: nil, Amount: model.NewMoney(decimal.NewFromInt(5000), "PLN")},
			},
			input: RecordFundingInput{
				UserID:     1,
				CategoryID: &catID1,
				YearMonth:  "2026-01",
				Amount:     model.NewMoney(decimal.NewFromInt(2000), "PLN"),
			},
		},
		"category exceeds overall budget": {
			existing: []model.Funding{
				{CategoryID: nil, Amount: model.NewMoney(decimal.NewFromInt(3000), "PLN")},
				{CategoryID: &catID1, Amount: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
			},
			input: RecordFundingInput{
				UserID:     1,
				CategoryID: &catID2,
				YearMonth:  "2026-01",
				Amount:     model.NewMoney(decimal.NewFromInt(2000), "PLN"),
			},
			wantErr: model.ErrCategoryBudgetsExceed,
		},
		"overall below category sum": {
			existing: []model.Funding{
				{CategoryID: &catID1, Amount: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
				{CategoryID: &catID2, Amount: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
			},
			input: RecordFundingInput{
				UserID:     1,
				CategoryID: nil,
				YearMonth:  "2026-01",
				Amount:     model.NewMoney(decimal.NewFromInt(3000), "PLN"),
			},
			wantErr: model.ErrCategoryBudgetsExceed,
		},
		"no overall budget set - category allowed": {
			existing: []model.Funding{
				{CategoryID: &catID1, Amount: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
			},
			input: RecordFundingInput{
				UserID:     1,
				CategoryID: &catID2,
				YearMonth:  "2026-01",
				Amount:     model.NewMoney(decimal.NewFromInt(5000), "PLN"),
			},
		},
		"updating existing category budget": {
			existing: []model.Funding{
				{CategoryID: nil, Amount: model.NewMoney(decimal.NewFromInt(5000), "PLN")},
				{CategoryID: &catID1, Amount: model.NewMoney(decimal.NewFromInt(1000), "PLN")},
				{CategoryID: &catID2, Amount: model.NewMoney(decimal.NewFromInt(1000), "PLN")},
			},
			input: RecordFundingInput{
				UserID:     1,
				CategoryID: &catID1,
				YearMonth:  "2026-01",
				Amount:     model.NewMoney(decimal.NewFromInt(3000), "PLN"),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(model.RoleEditor), nil
				},
			}
			fs := &mockFundingStore{
				ListByWorkspaceFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
					return tc.existing, nil
				},
				UpsertFunc: func(_ context.Context, f model.Funding) (model.Funding, error) {
					f.ID = 1
					return f, nil
				},
			}
			svc := NewFundingService(fs, ws)

			_, err := svc.Record(context.Background(), 1, 1, tc.input)
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

func TestFundingService_ListByRange(t *testing.T) {
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
					if tc.role == "" {
						return model.WorkspaceMember{}, model.ErrNotWorkspaceMember
					}
					return memberWithRole(tc.role), nil
				},
			}
			fs := &mockFundingStore{
				ListByWorkspaceFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
					return []model.Funding{
						{ID: 1, YearMonth: "2026-01"},
						{ID: 2, YearMonth: "2026-02"},
					}, nil
				},
			}
			svc := NewFundingService(fs, ws)

			_, err := svc.ListByRange(context.Background(), 1, 1, "2026-01", "2026-02")
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

func TestFundingService_Delete(t *testing.T) {
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
			fs := &mockFundingStore{
				DeleteFunc: func(_ context.Context, _ model.FundingID, _ model.WorkspaceID) error {
					return nil
				},
			}
			svc := NewFundingService(fs, ws)

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
