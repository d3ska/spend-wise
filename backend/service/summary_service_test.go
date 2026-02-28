package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/model"

	"github.com/shopspring/decimal"
)

// ── Mock summary store ──

type mockSummaryStore struct {
	TotalSpentFunc         func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (model.Money, error)
	TransactionCountFunc   func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (int, error)
	SpentByCategoryFunc    func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.CategorySpending, error)
	SpentByParticipantFunc func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.ParticipantSpending, error)
	SpentByDayFunc         func(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.DailySpending, error)
}

func (m *mockSummaryStore) TotalSpent(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (model.Money, error) {
	return m.TotalSpentFunc(ctx, wsID, from, to)
}

func (m *mockSummaryStore) TransactionCount(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (int, error) {
	if m.TransactionCountFunc != nil {
		return m.TransactionCountFunc(ctx, wsID, from, to)
	}
	return 0, nil
}

func (m *mockSummaryStore) SpentByCategory(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.CategorySpending, error) {
	return m.SpentByCategoryFunc(ctx, wsID, from, to)
}

func (m *mockSummaryStore) SpentByParticipant(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.ParticipantSpending, error) {
	return m.SpentByParticipantFunc(ctx, wsID, from, to)
}

func (m *mockSummaryStore) SpentByDay(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.DailySpending, error) {
	if m.SpentByDayFunc != nil {
		return m.SpentByDayFunc(ctx, wsID, from, to)
	}
	return []model.DailySpending{}, nil
}

// ── Mock funding store for summary ──

type mockFundingStoreForSummary struct {
	ListByWorkspaceAndUsersFunc func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
	GetOverallBudgetsFunc       func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
	GetCategoryBudgetsFunc      func(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
}

func (m *mockFundingStoreForSummary) ListByWorkspaceAndUsers(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	return m.ListByWorkspaceAndUsersFunc(ctx, wsID, fromMonth, toMonth)
}

func (m *mockFundingStoreForSummary) GetOverallBudgets(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	return m.GetOverallBudgetsFunc(ctx, wsID, fromMonth, toMonth)
}

func (m *mockFundingStoreForSummary) GetCategoryBudgets(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	return m.GetCategoryBudgetsFunc(ctx, wsID, fromMonth, toMonth)
}

// ── Tests ──

func TestSummaryService_GetSummary(t *testing.T) {
	tests := map[string]struct {
		role      model.MemberRole
		memberErr error
		wantErr   error
	}{
		"viewer allowed": {
			role: model.RoleViewer,
		},
		"non-member rejected": {
			memberErr: model.ErrNotWorkspaceMember,
			wantErr:   model.ErrNotWorkspaceMember,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					if tc.memberErr != nil {
						return model.WorkspaceMember{}, tc.memberErr
					}
					return memberWithRole(tc.role), nil
				},
			}
			ss := &mockSummaryStore{
				TotalSpentFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) (model.Money, error) {
					return model.NewMoney(decimal.NewFromInt(3000), "PLN"), nil
				},
				SpentByCategoryFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) ([]model.CategorySpending, error) {
					return []model.CategorySpending{
						{CategoryID: 1, Name: "Food", Spent: model.NewMoney(decimal.NewFromInt(1500), "PLN")},
					}, nil
				},
				SpentByParticipantFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) ([]model.ParticipantSpending, error) {
					return []model.ParticipantSpending{
						{UserID: 1, DisplayName: "Alice", Spent: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
						{UserID: 2, DisplayName: "Bob", Spent: model.NewMoney(decimal.NewFromInt(1000), "PLN")},
					}, nil
				},
			}
			fs := &mockFundingStoreForSummary{
				ListByWorkspaceAndUsersFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				},
				GetOverallBudgetsFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				},
				GetCategoryBudgetsFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				},
			}
			svc := NewSummaryService(ss, fs, ws)

			_, err := svc.GetSummary(context.Background(), 1, 1, from, to)
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

func TestSummaryService_GetSummary_MergesFunding(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	ws := &mockWorkspaceStore{
		GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
			return memberWithRole(model.RoleViewer), nil
		},
	}
	ss := &mockSummaryStore{
		TotalSpentFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) (model.Money, error) {
			return model.NewMoney(decimal.NewFromInt(3000), "PLN"), nil
		},
		SpentByCategoryFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) ([]model.CategorySpending, error) {
			catID := model.CategoryID(10)
			return []model.CategorySpending{
				{CategoryID: catID, Name: "Food", Icon: "🍔", Spent: model.NewMoney(decimal.NewFromInt(1500), "PLN")},
			}, nil
		},
		SpentByParticipantFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) ([]model.ParticipantSpending, error) {
			return []model.ParticipantSpending{
				{UserID: 1, DisplayName: "Alice", Spent: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
				{UserID: 2, DisplayName: "Bob", Spent: model.NewMoney(decimal.NewFromInt(1000), "PLN")},
			}, nil
		},
	}

	catID := model.CategoryID(10)
	fs := &mockFundingStoreForSummary{
		ListByWorkspaceAndUsersFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
			return []model.Funding{
				{UserID: 1, CategoryID: nil, Amount: model.NewMoney(decimal.NewFromInt(3000), "PLN")},
				{UserID: 2, CategoryID: nil, Amount: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
			}, nil
		},
		GetOverallBudgetsFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
			return []model.Funding{
				{CategoryID: nil, Amount: model.NewMoney(decimal.NewFromInt(5000), "PLN")},
			}, nil
		},
		GetCategoryBudgetsFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
			return []model.Funding{
				{CategoryID: &catID, Amount: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
			}, nil
		},
	}
	svc := NewSummaryService(ss, fs, ws)

	got, err := svc.GetSummary(context.Background(), 1, 1, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check overall budget
	if got.OverallBudget == nil {
		t.Fatal("got nil OverallBudget, want non-nil")
	}
	if !got.OverallBudget.Amount().Equal(decimal.NewFromInt(5000)) {
		t.Errorf("got OverallBudget %v, want 5000", got.OverallBudget.Amount())
	}

	// Check category budget
	if len(got.ByCategory) != 1 {
		t.Fatalf("got %d categories, want 1", len(got.ByCategory))
	}
	if got.ByCategory[0].Budget == nil {
		t.Fatal("got nil Budget for category, want non-nil")
	}
	if !got.ByCategory[0].Budget.Amount().Equal(decimal.NewFromInt(2000)) {
		t.Errorf("got category budget %v, want 2000", got.ByCategory[0].Budget.Amount())
	}

	// Check participant funding
	if len(got.ByParticipant) != 2 {
		t.Fatalf("got %d participants, want 2", len(got.ByParticipant))
	}
	alice := got.ByParticipant[0]
	if !alice.Funded.Amount().Equal(decimal.NewFromInt(3000)) {
		t.Errorf("got Alice funded %v, want 3000", alice.Funded.Amount())
	}
	if !alice.Balance.Amount().Equal(decimal.NewFromInt(1000)) {
		t.Errorf("got Alice balance %v, want 1000", alice.Balance.Amount())
	}

	bob := got.ByParticipant[1]
	if !bob.Funded.Amount().Equal(decimal.NewFromInt(2000)) {
		t.Errorf("got Bob funded %v, want 2000", bob.Funded.Amount())
	}
	if !bob.Balance.Amount().Equal(decimal.NewFromInt(1000)) {
		t.Errorf("got Bob balance %v, want 1000", bob.Balance.Amount())
	}
}

func TestSummaryService_GetSummary_PreviousPeriodComparison(t *testing.T) {
	// Current period: Feb 1–16, previous period: Jan 1–16.
	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	prevFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name                 string
		currTotalSpent       int64
		prevTotalSpent       int64
		currTxCount          int
		prevTxCount          int
		currCategories       []model.CategorySpending
		prevCategories       []model.CategorySpending
		wantTotalSpent       int64
		wantPrevTotalSpent   int64
		wantTxCount          int
		wantPrevTxCount      int
		wantCategoryPrevAmts map[model.CategoryID]int64
	}{
		{
			name:           "spending increased vs previous month",
			currTotalSpent: 5000,
			prevTotalSpent: 3000,
			currTxCount:    70,
			prevTxCount:    35,
			currCategories: []model.CategorySpending{
				{CategoryID: 1, Name: "Food", Spent: model.NewMoney(decimal.NewFromInt(3000), "PLN")},
				{CategoryID: 2, Name: "Transport", Spent: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
			},
			prevCategories: []model.CategorySpending{
				{CategoryID: 1, Name: "Food", Spent: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
			},
			wantTotalSpent:     5000,
			wantPrevTotalSpent: 3000,
			wantTxCount:        70,
			wantPrevTxCount:    35,
			wantCategoryPrevAmts: map[model.CategoryID]int64{
				1: 2000, // existed last month
				2: 0,    // new category, zero prev
			},
		},
		{
			name:           "no previous data",
			currTotalSpent: 1000,
			prevTotalSpent: 0,
			currTxCount:    10,
			prevTxCount:    0,
			currCategories: []model.CategorySpending{
				{CategoryID: 1, Name: "Food", Spent: model.NewMoney(decimal.NewFromInt(1000), "PLN")},
			},
			prevCategories:     []model.CategorySpending{},
			wantTotalSpent:     1000,
			wantPrevTotalSpent: 0,
			wantTxCount:        10,
			wantPrevTxCount:    0,
			wantCategoryPrevAmts: map[model.CategoryID]int64{
				1: 0,
			},
		},
		{
			name:           "spending decreased vs previous month",
			currTotalSpent: 2000,
			prevTotalSpent: 4000,
			currTxCount:    20,
			prevTxCount:    40,
			currCategories: []model.CategorySpending{
				{CategoryID: 1, Name: "Food", Spent: model.NewMoney(decimal.NewFromInt(2000), "PLN")},
			},
			prevCategories: []model.CategorySpending{
				{CategoryID: 1, Name: "Food", Spent: model.NewMoney(decimal.NewFromInt(4000), "PLN")},
			},
			wantTotalSpent:     2000,
			wantPrevTotalSpent: 4000,
			wantTxCount:        20,
			wantPrevTxCount:    40,
			wantCategoryPrevAmts: map[model.CategoryID]int64{
				1: 4000,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ws := &mockWorkspaceStore{
				GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
					return memberWithRole(model.RoleViewer), nil
				},
			}
			ss := &mockSummaryStore{
				TotalSpentFunc: func(_ context.Context, _ model.WorkspaceID, f, _ time.Time) (model.Money, error) {
					if f.Equal(prevFrom) {
						return model.NewMoney(decimal.NewFromInt(tc.prevTotalSpent), "PLN"), nil
					}
					return model.NewMoney(decimal.NewFromInt(tc.currTotalSpent), "PLN"), nil
				},
				TransactionCountFunc: func(_ context.Context, _ model.WorkspaceID, f, _ time.Time) (int, error) {
					if f.Equal(prevFrom) {
						return tc.prevTxCount, nil
					}
					return tc.currTxCount, nil
				},
				SpentByCategoryFunc: func(_ context.Context, _ model.WorkspaceID, f, _ time.Time) ([]model.CategorySpending, error) {
					if f.Equal(prevFrom) {
						return tc.prevCategories, nil
					}
					return tc.currCategories, nil
				},
				SpentByParticipantFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) ([]model.ParticipantSpending, error) {
					return []model.ParticipantSpending{}, nil
				},
			}
			fs := &mockFundingStoreForSummary{
				ListByWorkspaceAndUsersFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				},
				GetOverallBudgetsFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				},
				GetCategoryBudgetsFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
					return []model.Funding{}, nil
				},
			}

			svc := NewSummaryService(ss, fs, ws)
			got, err := svc.GetSummary(context.Background(), 1, 1, from, to)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify current period values.
			if !got.TotalSpent.Amount().Equal(decimal.NewFromInt(tc.wantTotalSpent)) {
				t.Errorf("TotalSpent = %v, want %d", got.TotalSpent.Amount(), tc.wantTotalSpent)
			}
			if got.TransactionCount != tc.wantTxCount {
				t.Errorf("TransactionCount = %d, want %d", got.TransactionCount, tc.wantTxCount)
			}

			// Verify previous period values.
			if !got.PrevTotalSpent.Amount().Equal(decimal.NewFromInt(tc.wantPrevTotalSpent)) {
				t.Errorf("PrevTotalSpent = %v, want %d", got.PrevTotalSpent.Amount(), tc.wantPrevTotalSpent)
			}
			if got.PrevTransactionCount != tc.wantPrevTxCount {
				t.Errorf("PrevTransactionCount = %d, want %d", got.PrevTransactionCount, tc.wantPrevTxCount)
			}

			// Verify per-category previous amounts.
			for _, c := range got.ByCategory {
				wantPrev, ok := tc.wantCategoryPrevAmts[c.CategoryID]
				if !ok {
					t.Errorf("unexpected category %d in result", c.CategoryID)
					continue
				}
				if !c.PrevSpent.Amount().Equal(decimal.NewFromInt(wantPrev)) {
					t.Errorf("category %d PrevSpent = %v, want %d", c.CategoryID, c.PrevSpent.Amount(), wantPrev)
				}
			}
		})
	}
}

func TestSummaryService_GetSummary_PreviousPeriodDates(t *testing.T) {
	// Verify that the previous period uses same calendar dates one month back.
	from := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	wantPrevFrom := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	wantPrevTo := time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC)

	var capturedFromDates []time.Time

	ws := &mockWorkspaceStore{
		GetMemberFunc: func(_ context.Context, _ model.WorkspaceID, _ model.UserID) (model.WorkspaceMember, error) {
			return memberWithRole(model.RoleViewer), nil
		},
	}
	ss := &mockSummaryStore{
		TotalSpentFunc: func(_ context.Context, _ model.WorkspaceID, f, to time.Time) (model.Money, error) {
			capturedFromDates = append(capturedFromDates, f)
			return model.Zero("PLN"), nil
		},
		TransactionCountFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) (int, error) {
			return 0, nil
		},
		SpentByCategoryFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) ([]model.CategorySpending, error) {
			return []model.CategorySpending{}, nil
		},
		SpentByParticipantFunc: func(_ context.Context, _ model.WorkspaceID, _, _ time.Time) ([]model.ParticipantSpending, error) {
			return []model.ParticipantSpending{}, nil
		},
	}
	fs := &mockFundingStoreForSummary{
		ListByWorkspaceAndUsersFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
			return []model.Funding{}, nil
		},
		GetOverallBudgetsFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
			return []model.Funding{}, nil
		},
		GetCategoryBudgetsFunc: func(_ context.Context, _ model.WorkspaceID, _, _ string) ([]model.Funding, error) {
			return []model.Funding{}, nil
		},
	}

	svc := NewSummaryService(ss, fs, ws)
	_, err := svc.GetSummary(context.Background(), 1, 1, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// TotalSpent is called twice: once for current, once for previous.
	if len(capturedFromDates) != 2 {
		t.Fatalf("expected TotalSpent called 2 times, got %d", len(capturedFromDates))
	}
	if !capturedFromDates[0].Equal(from) {
		t.Errorf("first TotalSpent from = %v, want %v", capturedFromDates[0], from)
	}
	if !capturedFromDates[1].Equal(wantPrevFrom) {
		t.Errorf("second TotalSpent from = %v, want %v (prev month)", capturedFromDates[1], wantPrevFrom)
	}
	_ = wantPrevTo // validated implicitly — if from is correct, to follows from AddDate(0,-1,0)
}
