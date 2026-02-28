package store

import (
	"context"
	"fmt"

	"backend/model"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FundingStore provides funding persistence operations.
type FundingStore struct {
	q *Queries
}

// NewFundingStore creates a new FundingStore.
func NewFundingStore(pool *pgxpool.Pool) *FundingStore {
	return &FundingStore{q: New(pool)}
}

// Upsert inserts or updates a funding record for the given workspace/user/month/category.
func (s *FundingStore) Upsert(ctx context.Context, f model.Funding) (model.Funding, error) {
	row, err := s.q.UpsertFunding(ctx, UpsertFundingParams{
		WorkspaceID: int64(f.WorkspaceID),
		UserID:      int64(f.UserID),
		YearMonth:   f.YearMonth,
		Amount:      decimalToNumeric(f.Amount.Amount()),
		Currency:    f.Amount.Currency(),
		CategoryID:  categoryIDToPgInt8(f.CategoryID),
	})
	if err != nil {
		return model.Funding{}, fmt.Errorf("upserting funding: %w", err)
	}
	return toModelFunding(row.ID, row.WorkspaceID, row.UserID, row.YearMonth, row.Amount, row.Currency, row.CategoryID, row.CreatedAt, row.UpdatedAt), nil
}

// ListByWorkspace returns fundings for a workspace within a year_month range.
func (s *FundingStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	rows, err := s.q.ListFundingsByWorkspace(ctx, ListFundingsByWorkspaceParams{
		WorkspaceID: int64(wsID),
		YearMonth:   fromMonth,
		YearMonth_2: toMonth,
	})
	if err != nil {
		return nil, fmt.Errorf("listing fundings: %w", err)
	}

	result := make([]model.Funding, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelFunding(row.ID, row.WorkspaceID, row.UserID, row.YearMonth, row.Amount, row.Currency, row.CategoryID, row.CreatedAt, row.UpdatedAt))
	}
	return result, nil
}

// ListByWorkspaceAndUsers returns fundings for a workspace grouped by user within a year_month range.
func (s *FundingStore) ListByWorkspaceAndUsers(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	rows, err := s.q.GetFundingsByWorkspaceAndUsers(ctx, GetFundingsByWorkspaceAndUsersParams{
		WorkspaceID: int64(wsID),
		YearMonth:   fromMonth,
		YearMonth_2: toMonth,
	})
	if err != nil {
		return nil, fmt.Errorf("listing fundings by users: %w", err)
	}

	result := make([]model.Funding, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelFunding(row.ID, row.WorkspaceID, row.UserID, row.YearMonth, row.Amount, row.Currency, row.CategoryID, row.CreatedAt, row.UpdatedAt))
	}
	return result, nil
}

// Delete removes a funding record by ID within a workspace.
func (s *FundingStore) Delete(ctx context.Context, id model.FundingID, wsID model.WorkspaceID) error {
	result, err := s.q.DeleteFunding(ctx, DeleteFundingParams{
		ID:          int64(id),
		WorkspaceID: int64(wsID),
	})
	if err != nil {
		return fmt.Errorf("deleting funding: %w", err)
	}
	if result.RowsAffected() == 0 {
		return model.ErrFundingNotFound
	}
	return nil
}

// GetOverallBudgets returns overall budget (category_id IS NULL) fundings for a workspace in a month range.
func (s *FundingStore) GetOverallBudgets(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	rows, err := s.q.GetOverallBudgetByWorkspace(ctx, GetOverallBudgetByWorkspaceParams{
		WorkspaceID: int64(wsID),
		YearMonth:   fromMonth,
		YearMonth_2: toMonth,
	})
	if err != nil {
		return nil, fmt.Errorf("getting overall budgets: %w", err)
	}

	result := make([]model.Funding, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelFunding(row.ID, row.WorkspaceID, row.UserID, row.YearMonth, row.Amount, row.Currency, row.CategoryID, row.CreatedAt, row.UpdatedAt))
	}
	return result, nil
}

// GetCategoryBudgets returns per-category budget fundings for a workspace in a month range.
func (s *FundingStore) GetCategoryBudgets(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	rows, err := s.q.GetCategoryBudgetsByWorkspace(ctx, GetCategoryBudgetsByWorkspaceParams{
		WorkspaceID: int64(wsID),
		YearMonth:   fromMonth,
		YearMonth_2: toMonth,
	})
	if err != nil {
		return nil, fmt.Errorf("getting category budgets: %w", err)
	}

	result := make([]model.Funding, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelFunding(row.ID, row.WorkspaceID, row.UserID, row.YearMonth, row.Amount, row.Currency, row.CategoryID, row.CreatedAt, row.UpdatedAt))
	}
	return result, nil
}

func toModelFunding(id int64, wsID int64, userID int64, yearMonth string, amount pgtype.Numeric, currency string, categoryID pgtype.Int8, createdAt, updatedAt pgtype.Timestamptz) model.Funding {
	f := model.Funding{
		ID:          model.FundingID(id),
		WorkspaceID: model.WorkspaceID(wsID),
		UserID:      model.UserID(userID),
		YearMonth:   yearMonth,
		Amount:      model.NewMoney(numericToDecimal(amount), currency),
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}
	if categoryID.Valid {
		catID := model.CategoryID(categoryID.Int64)
		f.CategoryID = &catID
	}
	return f
}

func categoryIDToPgInt8(catID *model.CategoryID) pgtype.Int8 {
	if catID == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: int64(*catID), Valid: true}
}
