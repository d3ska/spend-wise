package store

import (
	"context"
	"fmt"
	"time"

	"backend/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// SummaryStore provides aggregate spending queries for the summary endpoint.
// Uses manual pool.Query (not sqlc) because these return custom aggregate shapes.
type SummaryStore struct {
	pool *pgxpool.Pool
}

// NewSummaryStore creates a new SummaryStore.
func NewSummaryStore(pool *pgxpool.Pool) *SummaryStore {
	return &SummaryStore{pool: pool}
}

// TotalSpent returns the sum of all entry amounts for the given workspace and date range.
func (s *SummaryStore) TotalSpent(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (model.Money, error) {
	query := `
		SELECT COALESCE(SUM(e.amount), 0)
		FROM entries e
		JOIN transactions t ON t.id = e.transaction_id
		WHERE t.workspace_id = $1
		  AND t.date >= $2
		  AND t.date < $3
		  AND t.transaction_type = 'expense'`

	var total decimal.Decimal
	if err := s.pool.QueryRow(ctx, query, int64(wsID), from, to).Scan(&total); err != nil {
		return model.Money{}, fmt.Errorf("querying total spent: %w", err)
	}
	return model.NewMoney(total, model.DefaultCurrency), nil
}

// TransactionCount returns the number of expense transactions for the given workspace and date range.
func (s *SummaryStore) TransactionCount(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM transactions t
		WHERE t.workspace_id = $1
		  AND t.date >= $2
		  AND t.date < $3
		  AND t.transaction_type = 'expense'`

	var count int
	if err := s.pool.QueryRow(ctx, query, int64(wsID), from, to).Scan(&count); err != nil {
		return 0, fmt.Errorf("querying transaction count: %w", err)
	}
	return count, nil
}

// SpentByDay returns daily spending totals for the given workspace and date range.
func (s *SummaryStore) SpentByDay(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.DailySpending, error) {
	query := `
		SELECT t.date, COALESCE(SUM(e.amount), 0) AS spent
		FROM entries e
		JOIN transactions t ON t.id = e.transaction_id
		WHERE t.workspace_id = $1
		  AND t.date >= $2
		  AND t.date < $3
		  AND t.transaction_type = 'expense'
		GROUP BY t.date
		ORDER BY t.date`

	rows, err := s.pool.Query(ctx, query, int64(wsID), from, to)
	if err != nil {
		return nil, fmt.Errorf("querying spent by day: %w", err)
	}
	defer rows.Close()

	var result []model.DailySpending
	for rows.Next() {
		var (
			date  time.Time
			spent decimal.Decimal
		)
		if err := rows.Scan(&date, &spent); err != nil {
			return nil, fmt.Errorf("scanning daily spending: %w", err)
		}
		result = append(result, model.DailySpending{
			Date:  date,
			Spent: model.NewMoney(spent, model.DefaultCurrency),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating daily spending: %w", err)
	}

	if result == nil {
		result = []model.DailySpending{}
	}
	return result, nil
}

// SpentByCategory returns spending per category for the given workspace and date range.
func (s *SummaryStore) SpentByCategory(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.CategorySpending, error) {
	query := `
		SELECT c.id, c.name, c.icon, c.slug, COALESCE(SUM(e.amount), 0) AS spent
		FROM entries e
		JOIN transactions t ON t.id = e.transaction_id
		JOIN categories c ON c.id = e.category_id
		WHERE t.workspace_id = $1
		  AND t.date >= $2
		  AND t.date < $3
		  AND t.transaction_type = 'expense'
		GROUP BY c.id, c.name, c.icon, c.slug
		ORDER BY spent DESC`

	rows, err := s.pool.Query(ctx, query, int64(wsID), from, to)
	if err != nil {
		return nil, fmt.Errorf("querying spent by category: %w", err)
	}
	defer rows.Close()

	var result []model.CategorySpending
	for rows.Next() {
		var (
			catID int64
			name  string
			icon  string
			slug  *string
			spent decimal.Decimal
		)
		if err := rows.Scan(&catID, &name, &icon, &slug, &spent); err != nil {
			return nil, fmt.Errorf("scanning category spending: %w", err)
		}
		result = append(result, model.CategorySpending{
			CategoryID: model.CategoryID(catID),
			Name:       name,
			Icon:       icon,
			Slug:       slug,
			Spent:      model.NewMoney(spent, model.DefaultCurrency),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating category spending: %w", err)
	}

	if result == nil {
		result = []model.CategorySpending{}
	}
	return result, nil
}

// SpentByParticipant returns spending per participant for the given workspace and date range.
// Entries with NULL participant_id are excluded.
func (s *SummaryStore) SpentByParticipant(ctx context.Context, wsID model.WorkspaceID, from, to time.Time) ([]model.ParticipantSpending, error) {
	query := `
		SELECT e.participant_id, u.display_name, COALESCE(SUM(e.amount), 0) AS spent
		FROM entries e
		JOIN transactions t ON t.id = e.transaction_id
		JOIN users u ON u.id = e.participant_id
		WHERE t.workspace_id = $1
		  AND t.date >= $2
		  AND t.date < $3
		  AND e.participant_id IS NOT NULL
		  AND t.transaction_type = 'expense'
		GROUP BY e.participant_id, u.display_name
		ORDER BY spent DESC`

	rows, err := s.pool.Query(ctx, query, int64(wsID), from, to)
	if err != nil {
		return nil, fmt.Errorf("querying spent by participant: %w", err)
	}
	defer rows.Close()

	var result []model.ParticipantSpending
	for rows.Next() {
		var (
			userID      int64
			displayName string
			spent       decimal.Decimal
		)
		if err := rows.Scan(&userID, &displayName, &spent); err != nil {
			return nil, fmt.Errorf("scanning participant spending: %w", err)
		}
		spentMoney := model.NewMoney(spent, model.DefaultCurrency)
		result = append(result, model.ParticipantSpending{
			UserID:      model.UserID(userID),
			DisplayName: displayName,
			Spent:       spentMoney,
			Funded:      model.Zero(model.DefaultCurrency),
			Balance:     model.Zero(model.DefaultCurrency).Sub(spentMoney),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating participant spending: %w", err)
	}

	if result == nil {
		result = []model.ParticipantSpending{}
	}
	return result, nil
}
