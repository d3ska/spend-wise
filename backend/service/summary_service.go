package service

import (
	"context"
	"fmt"
	"time"

	"backend/model"
)

// SummaryStoreIface defines the summary store methods used by SummaryService.
type SummaryStoreIface interface {
	TotalSpent(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (model.Money, error)
	TransactionCount(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (int, error)
	SpentByCategory(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.CategorySpending, error)
	SpentByParticipant(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.ParticipantSpending, error)
	SpentByDay(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.DailySpending, error)
}

// FundingStoreForSummaryIface defines the funding store methods used by SummaryService.
type FundingStoreForSummaryIface interface {
	ListByWorkspaceAndUsers(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
	GetOverallBudgets(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
	GetCategoryBudgets(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
}

// SummaryService handles spending summary operations with permission checks.
type SummaryService struct {
	summaryStore SummaryStoreIface
	fundingStore FundingStoreForSummaryIface
	wsStore      WorkspaceStoreIface
}

// NewSummaryService creates a new SummaryService.
func NewSummaryService(summaryStore SummaryStoreIface, fundingStore FundingStoreForSummaryIface, wsStore WorkspaceStoreIface) *SummaryService {
	return &SummaryService{
		summaryStore: summaryStore,
		fundingStore: fundingStore,
		wsStore:      wsStore,
	}
}

// GetSummary returns a spending summary for the workspace in the given date range.
// Requires viewer+ permission.
func (s *SummaryService) GetSummary(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, from, to time.Time) (model.SpendingSummary, error) {
	member, err := s.wsStore.GetMember(ctx, wsID, userID)
	if err != nil {
		return model.SpendingSummary{}, err
	}
	if member.Role.Level() < model.RoleViewer.Level() {
		return model.SpendingSummary{}, model.ErrInsufficientPermission
	}

	totalSpent, err := s.summaryStore.TotalSpent(ctx, wsID, from, to)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting total spent: %w", err)
	}

	txCount, err := s.summaryStore.TransactionCount(ctx, wsID, from, to)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting transaction count: %w", err)
	}

	// Previous period: same calendar dates in the prior month.
	prevFrom := from.AddDate(0, -1, 0)
	prevTo := to.AddDate(0, -1, 0)

	prevTotalSpent, err := s.summaryStore.TotalSpent(ctx, wsID, prevFrom, prevTo)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting prev total spent: %w", err)
	}

	prevTxCount, err := s.summaryStore.TransactionCount(ctx, wsID, prevFrom, prevTo)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting prev transaction count: %w", err)
	}

	dailyTrend, err := s.summaryStore.SpentByDay(ctx, wsID, from, to)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting daily trend: %w", err)
	}

	byCategory, err := s.summaryStore.SpentByCategory(ctx, wsID, from, to)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting spent by category: %w", err)
	}

	prevByCategory, err := s.summaryStore.SpentByCategory(ctx, wsID, prevFrom, prevTo)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting prev spent by category: %w", err)
	}
	prevSpentByCat := make(map[model.CategoryID]model.Money, len(prevByCategory))
	for _, c := range prevByCategory {
		prevSpentByCat[c.CategoryID] = c.Spent
	}
	for i, c := range byCategory {
		if prev, ok := prevSpentByCat[c.CategoryID]; ok {
			byCategory[i].PrevSpent = prev
		} else {
			byCategory[i].PrevSpent = model.Zero(model.DefaultCurrency)
		}
	}

	byParticipant, err := s.summaryStore.SpentByParticipant(ctx, wsID, from, to)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting spent by participant: %w", err)
	}

	// Merge funding data into participant spending.
	fromMonth := from.Format("2006-01")
	toMonth := to.AddDate(0, 0, -1).Format("2006-01") // inclusive end month
	fundings, err := s.fundingStore.ListByWorkspaceAndUsers(ctx, wsID, fromMonth, toMonth)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting fundings: %w", err)
	}

	// Sum fundings per user.
	fundedByUser := make(map[model.UserID]model.Money)
	for _, f := range fundings {
		if existing, ok := fundedByUser[f.UserID]; ok {
			fundedByUser[f.UserID] = existing.Add(f.Amount)
		} else {
			fundedByUser[f.UserID] = f.Amount
		}
	}

	// Update ParticipantSpending with real funded amounts and balance.
	for i, p := range byParticipant {
		if funded, ok := fundedByUser[p.UserID]; ok {
			byParticipant[i].Funded = funded
			byParticipant[i].Balance = funded.Sub(p.Spent)
		}
	}

	// Fetch overall budget (category_id IS NULL) for the date range.
	overallBudgets, err := s.fundingStore.GetOverallBudgets(ctx, wsID, fromMonth, toMonth)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting overall budgets: %w", err)
	}
	var overallBudget *model.Money
	if len(overallBudgets) > 0 {
		sum := overallBudgets[0].Amount
		for _, b := range overallBudgets[1:] {
			sum = sum.Add(b.Amount)
		}
		overallBudget = &sum
	}

	// Fetch per-category budgets for the date range.
	catBudgets, err := s.fundingStore.GetCategoryBudgets(ctx, wsID, fromMonth, toMonth)
	if err != nil {
		return model.SpendingSummary{}, fmt.Errorf("getting category budgets: %w", err)
	}
	// Sum budgets per category across months.
	budgetByCategory := make(map[model.CategoryID]model.Money)
	for _, b := range catBudgets {
		if b.CategoryID == nil {
			continue
		}
		catID := *b.CategoryID
		if existing, ok := budgetByCategory[catID]; ok {
			budgetByCategory[catID] = existing.Add(b.Amount)
		} else {
			budgetByCategory[catID] = b.Amount
		}
	}
	// Attach budget to each CategorySpending.
	for i, c := range byCategory {
		if budget, ok := budgetByCategory[c.CategoryID]; ok {
			byCategory[i].Budget = &budget
		}
	}

	return model.SpendingSummary{
		WorkspaceID:          wsID,
		From:                 from,
		To:                   to,
		TotalSpent:           totalSpent,
		TransactionCount:     txCount,
		PrevTotalSpent:       prevTotalSpent,
		PrevTransactionCount: prevTxCount,
		OverallBudget:        overallBudget,
		ByCategory:           byCategory,
		ByParticipant:        byParticipant,
		DailyTrend:           dailyTrend,
	}, nil
}
