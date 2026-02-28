package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"backend/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// RuleStore provides categorization rule persistence operations.
type RuleStore struct {
	q *Queries
}

// NewRuleStore creates a new RuleStore.
func NewRuleStore(pool *pgxpool.Pool) *RuleStore {
	return &RuleStore{q: New(pool)}
}

// GetByID returns a single categorization rule by ID.
func (s *RuleStore) GetByID(ctx context.Context, id model.RuleID) (model.CategorizationRule, error) {
	row, err := s.q.GetRuleByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CategorizationRule{}, model.ErrRuleNotFound
		}
		return model.CategorizationRule{}, fmt.Errorf("getting rule: %w", err)
	}
	return toModelRule(row), nil
}

// Create inserts a new categorization rule.
func (s *RuleStore) Create(ctx context.Context, r model.CategorizationRule) (model.CategorizationRule, error) {
	row, err := s.q.InsertRule(ctx, InsertRuleParams{
		Scope:            RuleScope(r.Scope),
		OwnerID:          userIDPtrToInt8(r.OwnerID),
		WorkspaceID:      workspaceIDPtrToInt8(r.WorkspaceID),
		MatchPattern:     r.MatchPattern,
		TargetCategoryID: int64(r.TargetCategoryID),
		Priority:         int32(r.Priority),
		Enabled:          r.Enabled,
		AmountMin:        decimalPtrToNumeric(r.AmountMin),
		AmountMax:        decimalPtrToNumeric(r.AmountMax),
		BankAccountID:    bankAccountIDToPgInt8(r.BankAccountID),
		CounterpartyIban: stringPtrToText(r.CounterpartyIBAN),
	})
	if err != nil {
		return model.CategorizationRule{}, fmt.Errorf("inserting rule: %w", err)
	}
	return toModelRule(row), nil
}

// Update modifies a rule's pattern, target category, priority, amount range, bank account filter, and counterparty IBAN.
func (s *RuleStore) Update(ctx context.Context, id model.RuleID, pattern string, targetCatID model.CategoryID, priority int, amountMin, amountMax *decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (model.CategorizationRule, error) {
	row, err := s.q.UpdateRule(ctx, UpdateRuleParams{
		ID:               int64(id),
		MatchPattern:     pattern,
		TargetCategoryID: int64(targetCatID),
		Priority:         int32(priority),
		AmountMin:        decimalPtrToNumeric(amountMin),
		AmountMax:        decimalPtrToNumeric(amountMax),
		BankAccountID:    bankAccountIDToPgInt8(bankAccountID),
		CounterpartyIban: stringPtrToText(counterpartyIBAN),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CategorizationRule{}, model.ErrRuleNotFound
		}
		return model.CategorizationRule{}, fmt.Errorf("updating rule: %w", err)
	}
	return toModelRule(row), nil
}

// Delete removes a rule by ID.
func (s *RuleStore) Delete(ctx context.Context, id model.RuleID) error {
	if err := s.q.DeleteRule(ctx, int64(id)); err != nil {
		return fmt.Errorf("deleting rule: %w", err)
	}
	return nil
}

// ListByWorkspace returns all rules for a workspace, ordered by priority DESC.
func (s *RuleStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]model.CategorizationRule, error) {
	rows, err := s.q.ListRulesByWorkspace(ctx, pgtype.Int8{Int64: int64(wsID), Valid: true})
	if err != nil {
		return nil, fmt.Errorf("listing rules: %w", err)
	}

	result := make([]model.CategorizationRule, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelRule(row))
	}
	return result, nil
}

// ToggleEnabled sets the enabled flag for a rule.
func (s *RuleStore) ToggleEnabled(ctx context.Context, id model.RuleID, enabled bool) (model.CategorizationRule, error) {
	row, err := s.q.ToggleRuleEnabled(ctx, ToggleRuleEnabledParams{
		ID:      int64(id),
		Enabled: enabled,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CategorizationRule{}, model.ErrRuleNotFound
		}
		return model.CategorizationRule{}, fmt.Errorf("toggling rule enabled: %w", err)
	}
	return toModelRule(row), nil
}

// ResolveForTransaction fetches all enabled rules for the given workspace/user
// in precedence order, and returns the first one whose pattern matches (case-insensitive),
// whose amount range includes the transaction amount, and whose bank account filter matches.
func (s *RuleStore) ResolveForTransaction(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, description string, amount decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (*model.CategoryID, error) {
	rows, err := s.q.ListEnabledRulesForResolution(ctx, ListEnabledRulesForResolutionParams{
		WorkspaceID: pgtype.Int8{Int64: int64(wsID), Valid: true},
		OwnerID:     pgtype.Int8{Int64: int64(userID), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("listing rules for resolution: %w", err)
	}

	absAmount := amount.Abs()
	for _, row := range rows {
		re, err := regexp.Compile(row.MatchPattern)
		if err != nil {
			slog.Warn("skipping rule with invalid regex pattern", "rule_id", row.ID, "error", err)
			continue
		}
		if !re.MatchString(description) {
			continue
		}

		rule := toModelRule(row)
		if !rule.MatchesAmount(absAmount) {
			continue
		}
		if !rule.MatchesBankAccount(bankAccountID) {
			continue
		}
		if !rule.MatchesCounterpartyIBAN(counterpartyIBAN) {
			continue
		}

		catID := model.CategoryID(row.TargetCategoryID)
		return &catID, nil
	}
	return nil, nil
}

func workspaceIDPtrToInt8(wsID *model.WorkspaceID) pgtype.Int8 {
	if wsID == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: int64(*wsID), Valid: true}
}

func int8ToWorkspaceIDPtr(i pgtype.Int8) *model.WorkspaceID {
	if !i.Valid {
		return nil
	}
	wsID := model.WorkspaceID(i.Int64)
	return &wsID
}

func numericToDecimalPtr(n pgtype.Numeric) *decimal.Decimal {
	if !n.Valid {
		return nil
	}
	d := decimal.NewFromBigInt(n.Int, n.Exp)
	return &d
}

func decimalPtrToNumeric(d *decimal.Decimal) pgtype.Numeric {
	if d == nil {
		return pgtype.Numeric{}
	}
	var n pgtype.Numeric
	_ = n.Scan(d.String())
	return n
}

func toModelRule(row CategorizationRule) model.CategorizationRule {
	return model.CategorizationRule{
		ID:               model.RuleID(row.ID),
		Scope:            model.RuleScope(row.Scope),
		OwnerID:          int8ToUserIDPtr(row.OwnerID),
		WorkspaceID:      int8ToWorkspaceIDPtr(row.WorkspaceID),
		MatchPattern:     row.MatchPattern,
		TargetCategoryID: model.CategoryID(row.TargetCategoryID),
		Priority:         int(row.Priority),
		Enabled:          row.Enabled,
		AmountMin:        numericToDecimalPtr(row.AmountMin),
		AmountMax:        numericToDecimalPtr(row.AmountMax),
		BankAccountID:    int8ToBankAccountIDPtr(row.BankAccountID),
		CounterpartyIBAN: textToStringPtr(row.CounterpartyIban),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}
