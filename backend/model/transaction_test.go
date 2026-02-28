package model

import (
	"errors"
	"testing"
	"time"
)

func TestTransaction_ValidateEntries(t *testing.T) {
	pln := func(s string) Money { return NewMoney(dec(s), "PLN") }
	validDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		tx      Transaction
		wantErr error
	}{
		"no entries": {
			tx:      Transaction{TotalAmount: pln("100"), Date: validDate, Entries: nil},
			wantErr: ErrTransactionNoEntries,
		},
		"empty entries slice": {
			tx:      Transaction{TotalAmount: pln("100"), Date: validDate, Entries: []Entry{}},
			wantErr: ErrTransactionNoEntries,
		},
		"sum mismatch": {
			tx: Transaction{
				TotalAmount: pln("100"),
				Date:        validDate,
				Entries: []Entry{
					{Amount: pln("60")},
					{Amount: pln("30")},
				},
			},
			wantErr: ErrEntrySumMismatch,
		},
		"valid single entry": {
			tx: Transaction{
				TotalAmount: pln("50.00"),
				Date:        validDate,
				Entries:     []Entry{{Amount: pln("50.00")}},
			},
			wantErr: nil,
		},
		"valid multiple entries": {
			tx: Transaction{
				TotalAmount: pln("100.00"),
				Date:        validDate,
				Entries: []Entry{
					{Amount: pln("40.00")},
					{Amount: pln("35.50")},
					{Amount: pln("24.50")},
				},
			},
			wantErr: nil,
		},
		"amount too large": {
			tx: Transaction{
				TotalAmount: pln("9999999999"),
				Date:        validDate,
				Entries:     []Entry{{Amount: pln("9999999999")}},
			},
			wantErr: ErrTransactionAmountTooLarge,
		},
		"date too far in past": {
			tx: Transaction{
				TotalAmount: pln("50.00"),
				Date:        time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC),
				Entries:     []Entry{{Amount: pln("50.00")}},
			},
			wantErr: ErrTransactionDateTooFarInPast,
		},
		"date too far in future": {
			tx: Transaction{
				TotalAmount: pln("50.00"),
				Date:        time.Now().AddDate(2, 0, 0),
				Entries:     []Entry{{Amount: pln("50.00")}},
			},
			wantErr: ErrTransactionDateTooFarInFuture,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.tx.ValidateEntries()
			if tc.wantErr == nil {
				if err != nil {
					t.Errorf("got %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
