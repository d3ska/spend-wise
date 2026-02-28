package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// TransactionID is a typed wrapper for transaction identifiers.
type TransactionID int64

// EntryID is a typed wrapper for entry identifiers.
type EntryID int64

// TransactionSource represents how a transaction was created.
type TransactionSource string

const (
	SourceManual TransactionSource = "manual"
	SourceImport TransactionSource = "import"
	SourceBank   TransactionSource = "bank"
)

// TransactionType distinguishes expenses from income.
type TransactionType string

const (
	TypeExpense  TransactionType = "expense"
	TypeIncome   TransactionType = "income"
	TypeTransfer TransactionType = "transfer"
)

// Transaction represents a financial transaction with one or more entries.
type Transaction struct {
	ID               TransactionID
	WorkspaceID      WorkspaceID
	CreatedBy        UserID
	TotalAmount      Money
	Description      string
	Date             time.Time
	Fingerprint      *string
	Source           TransactionSource
	Type             TransactionType
	BankAccountID    *BankAccountID
	IBAN             *string
	CounterpartyIBAN *string
	BankName         string // derived: institution name from bank_accounts → bank_connections
	Notes            string
	Entries          []Entry
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// maxTransactionAmount is a reasonable upper bound to prevent overflow/abuse.
// 999,999,999.99 — just under one billion in the transaction currency.
var maxTransactionAmount = decimal.NewFromInt(999_999_999)

// ValidateEntries checks that the transaction has at least one entry,
// that entry amounts sum to the total amount, and that amounts and dates
// are within reasonable bounds.
func (t Transaction) ValidateEntries() error {
	if len(t.Entries) == 0 {
		return ErrTransactionNoEntries
	}

	// Check total amount is within reasonable bounds.
	if t.TotalAmount.Amount().Abs().GreaterThan(maxTransactionAmount) {
		return ErrTransactionAmountTooLarge
	}

	// Check date bounds: not before year 2000, not more than 1 year in the future.
	minDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDate := time.Now().AddDate(1, 0, 0)
	if t.Date.Before(minDate) {
		return ErrTransactionDateTooFarInPast
	}
	if t.Date.After(maxDate) {
		return ErrTransactionDateTooFarInFuture
	}

	sum := Zero(t.TotalAmount.Currency())
	for _, e := range t.Entries {
		sum = sum.Add(e.Amount)
	}

	if !sum.Equal(t.TotalAmount) {
		return ErrEntrySumMismatch
	}
	return nil
}

// Entry represents a line item within a transaction, assigning an amount
// to a category and optionally a participant.
type Entry struct {
	ID            EntryID
	TransactionID TransactionID
	CategoryID    CategoryID
	ParticipantID *UserID
	Amount        Money
	Note          string
}
