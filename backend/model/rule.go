package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// RuleID is a typed wrapper for categorization rule identifiers.
type RuleID int64

// RuleScope represents the scope level of a categorization rule.
type RuleScope string

const (
	ScopeSystem    RuleScope = "system"
	ScopeUser      RuleScope = "user"
	ScopeWorkspace RuleScope = "workspace"
)

// CategorizationRule defines a pattern-based rule for auto-categorizing transactions.
type CategorizationRule struct {
	ID               RuleID
	Scope            RuleScope
	OwnerID          *UserID
	WorkspaceID      *WorkspaceID
	MatchPattern     string
	TargetCategoryID CategoryID
	Priority         int
	Enabled          bool
	AmountMin        *decimal.Decimal
	AmountMax        *decimal.Decimal
	BankAccountID    *BankAccountID
	CounterpartyIBAN *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ScopePrecedence returns the precedence level for the rule's scope.
// Higher values take priority: workspace (3) > user (2) > system (1).
func (r CategorizationRule) ScopePrecedence() int {
	switch r.Scope {
	case ScopeWorkspace:
		return 3
	case ScopeUser:
		return 2
	case ScopeSystem:
		return 1
	default:
		return 0
	}
}

// MatchesBankAccount returns true if the transaction's bank account matches
// the rule's filter. A nil BankAccountID on the rule means "match all accounts".
func (r CategorizationRule) MatchesBankAccount(txBankAccountID *BankAccountID) bool {
	if r.BankAccountID == nil {
		return true
	}
	if txBankAccountID == nil {
		return false
	}
	return *r.BankAccountID == *txBankAccountID
}

// MatchesCounterpartyIBAN returns true if the transaction's counterparty IBAN matches
// the rule's filter. A nil CounterpartyIBAN on the rule means "match all counterparties".
func (r CategorizationRule) MatchesCounterpartyIBAN(txCounterpartyIBAN *string) bool {
	if r.CounterpartyIBAN == nil {
		return true
	}
	if txCounterpartyIBAN == nil {
		return false
	}
	return *r.CounterpartyIBAN == *txCounterpartyIBAN
}

// MatchesAmount returns true if the given absolute amount falls within
// the rule's amount range. Nil bounds are treated as unbounded.
func (r CategorizationRule) MatchesAmount(absAmount decimal.Decimal) bool {
	if r.AmountMin != nil && absAmount.LessThan(*r.AmountMin) {
		return false
	}
	if r.AmountMax != nil && absAmount.GreaterThan(*r.AmountMax) {
		return false
	}
	return true
}
