package model

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCategorizationRule_ScopePrecedence(t *testing.T) {
	tests := map[string]struct {
		scope RuleScope
		want  int
	}{
		"workspace": {scope: ScopeWorkspace, want: 3},
		"user":      {scope: ScopeUser, want: 2},
		"system":    {scope: ScopeSystem, want: 1},
		"unknown":   {scope: RuleScope("other"), want: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := CategorizationRule{Scope: tc.scope}
			got := r.ScopePrecedence()
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func decPtr(v string) *decimal.Decimal {
	d, _ := decimal.NewFromString(v)
	return &d
}

func TestCategorizationRule_MatchesAmount(t *testing.T) {
	tests := map[string]struct {
		min    *decimal.Decimal
		max    *decimal.Decimal
		amount string
		want   bool
	}{
		"nil bounds match any": {
			min: nil, max: nil, amount: "999.99", want: true,
		},
		"only min — above": {
			min: decPtr("100"), max: nil, amount: "150", want: true,
		},
		"only min — below": {
			min: decPtr("100"), max: nil, amount: "50", want: false,
		},
		"only min — exact": {
			min: decPtr("100"), max: nil, amount: "100", want: true,
		},
		"only max — below": {
			min: nil, max: decPtr("200"), amount: "150", want: true,
		},
		"only max — above": {
			min: nil, max: decPtr("200"), amount: "250", want: false,
		},
		"only max — exact": {
			min: nil, max: decPtr("200"), amount: "200", want: true,
		},
		"within range": {
			min: decPtr("100"), max: decPtr("300"), amount: "200", want: true,
		},
		"below range": {
			min: decPtr("100"), max: decPtr("300"), amount: "50", want: false,
		},
		"above range": {
			min: decPtr("100"), max: decPtr("300"), amount: "350", want: false,
		},
		"exact min boundary": {
			min: decPtr("100"), max: decPtr("300"), amount: "100", want: true,
		},
		"exact max boundary": {
			min: decPtr("100"), max: decPtr("300"), amount: "300", want: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := CategorizationRule{AmountMin: tc.min, AmountMax: tc.max}
			amount, _ := decimal.NewFromString(tc.amount)
			got := r.MatchesAmount(amount)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func bankAcctIDPtr(id int64) *BankAccountID {
	baID := BankAccountID(id)
	return &baID
}

func TestCategorizationRule_MatchesBankAccount(t *testing.T) {
	tests := map[string]struct {
		ruleBankAccountID *BankAccountID
		txBankAccountID   *BankAccountID
		want              bool
	}{
		"nil rule matches nil tx": {
			ruleBankAccountID: nil, txBankAccountID: nil, want: true,
		},
		"nil rule matches any tx": {
			ruleBankAccountID: nil, txBankAccountID: bankAcctIDPtr(5), want: true,
		},
		"set rule — nil tx does not match": {
			ruleBankAccountID: bankAcctIDPtr(5), txBankAccountID: nil, want: false,
		},
		"matching IDs": {
			ruleBankAccountID: bankAcctIDPtr(5), txBankAccountID: bankAcctIDPtr(5), want: true,
		},
		"non-matching IDs": {
			ruleBankAccountID: bankAcctIDPtr(5), txBankAccountID: bankAcctIDPtr(10), want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := CategorizationRule{BankAccountID: tc.ruleBankAccountID}
			got := r.MatchesBankAccount(tc.txBankAccountID)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestCategorizationRule_MatchesCounterpartyIBAN(t *testing.T) {
	tests := map[string]struct {
		ruleCounterpartyIBAN *string
		txCounterpartyIBAN   *string
		want                 bool
	}{
		"nil rule matches nil tx": {
			ruleCounterpartyIBAN: nil, txCounterpartyIBAN: nil, want: true,
		},
		"nil rule matches any tx": {
			ruleCounterpartyIBAN: nil, txCounterpartyIBAN: strPtr("PL1234567890"), want: true,
		},
		"set rule — nil tx does not match": {
			ruleCounterpartyIBAN: strPtr("PL1234567890"), txCounterpartyIBAN: nil, want: false,
		},
		"matching IBANs": {
			ruleCounterpartyIBAN: strPtr("PL1234567890"), txCounterpartyIBAN: strPtr("PL1234567890"), want: true,
		},
		"non-matching IBANs": {
			ruleCounterpartyIBAN: strPtr("PL1234567890"), txCounterpartyIBAN: strPtr("DE9876543210"), want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := CategorizationRule{CounterpartyIBAN: tc.ruleCounterpartyIBAN}
			got := r.MatchesCounterpartyIBAN(tc.txCounterpartyIBAN)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
