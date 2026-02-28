package model

import "time"

// FundingID is a typed wrapper for funding identifiers.
type FundingID int64

// Funding represents a budget target (overall or per-category) for a workspace for a given month.
type Funding struct {
	ID          FundingID
	WorkspaceID WorkspaceID
	UserID      UserID
	CategoryID  *CategoryID // nil = overall budget, non-nil = per-category budget
	YearMonth   string      // "YYYY-MM" format
	Amount      Money
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
