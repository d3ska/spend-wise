package model

import "time"

// SpendingSummary is a read model (never stored) representing spending
// aggregates for a workspace within a date range.
type SpendingSummary struct {
	WorkspaceID          WorkspaceID
	From                 time.Time
	To                   time.Time
	TotalSpent           Money
	TransactionCount     int
	PrevTotalSpent       Money
	PrevTransactionCount int
	OverallBudget        *Money
	ByCategory           []CategorySpending
	ByParticipant        []ParticipantSpending
	DailyTrend           []DailySpending
}

// DailySpending holds the total spent for a single day.
type DailySpending struct {
	Date  time.Time
	Spent Money
}

// CategorySpending holds the total spent for a single category.
type CategorySpending struct {
	CategoryID CategoryID
	Name       string
	Icon       string
	Slug       *string
	Spent      Money
	PrevSpent  Money
	Budget     *Money
}

// ParticipantSpending holds spending and funding totals for a single participant.
type ParticipantSpending struct {
	UserID      UserID
	DisplayName string
	Spent       Money
	Funded      Money
	Balance     Money
}
