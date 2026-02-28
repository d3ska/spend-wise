package model

import (
	"testing"

	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestNewMoney_DefaultCurrency(t *testing.T) {
	m := NewMoney(decimal.NewFromInt(100), "")
	if m.Currency() != "PLN" {
		t.Errorf("got currency %q, want PLN", m.Currency())
	}
}

func TestNewMoney_ExplicitCurrency(t *testing.T) {
	m := NewMoney(decimal.NewFromInt(100), "EUR")
	if m.Currency() != "EUR" {
		t.Errorf("got currency %q, want EUR", m.Currency())
	}
}

func TestNewMoneyFromString(t *testing.T) {
	tests := []struct {
		name    string
		amount  string
		cur     string
		want    string
		wantErr bool
	}{
		{"valid", "10.50", "PLN", "10.50 PLN", false},
		{"integer", "100", "EUR", "100.00 EUR", false},
		{"invalid", "not-a-number", "PLN", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMoneyFromString(tt.amount, tt.cur)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestMoney_Add(t *testing.T) {
	tests := []struct {
		name string
		a, b Money
		want Money
	}{
		{"zero plus zero", Zero("PLN"), Zero("PLN"), Zero("PLN")},
		{"positive sum", NewMoney(dec("10.50"), "PLN"), NewMoney(dec("3.25"), "PLN"), NewMoney(dec("13.75"), "PLN")},
		{"add to zero", Zero("PLN"), NewMoney(dec("5.00"), "PLN"), NewMoney(dec("5.00"), "PLN")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Add(tt.b)
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMoney_Sub(t *testing.T) {
	tests := []struct {
		name string
		a, b Money
		want Money
	}{
		{"positive result", NewMoney(dec("10.00"), "PLN"), NewMoney(dec("3.25"), "PLN"), NewMoney(dec("6.75"), "PLN")},
		{"zero result", NewMoney(dec("5.00"), "PLN"), NewMoney(dec("5.00"), "PLN"), Zero("PLN")},
		{"negative result", NewMoney(dec("3.00"), "PLN"), NewMoney(dec("5.00"), "PLN"), NewMoney(dec("-2.00"), "PLN")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Sub(tt.b)
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMoney_ArithmeticDoesNotMutate(t *testing.T) {
	a := NewMoney(dec("10.00"), "PLN")
	b := NewMoney(dec("3.00"), "PLN")

	a.Add(b)
	if !a.Amount().Equal(dec("10.00")) {
		t.Errorf("Add mutated receiver: got %v", a.Amount())
	}

	a.Sub(b)
	if !a.Amount().Equal(dec("10.00")) {
		t.Errorf("Sub mutated receiver: got %v", a.Amount())
	}
}

func TestMoney_CurrencyMismatchPanics(t *testing.T) {
	pln := NewMoney(dec("10.00"), "PLN")
	eur := NewMoney(dec("5.00"), "EUR")

	ops := []struct {
		name string
		fn   func()
	}{
		{"Add", func() { pln.Add(eur) }},
		{"Sub", func() { pln.Sub(eur) }},
		{"GreaterThan", func() { pln.GreaterThan(eur) }},
		{"LessThanOrEqual", func() { pln.LessThanOrEqual(eur) }},
	}
	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Error("expected panic, got none")
				}
			}()
			op.fn()
		})
	}
}

func TestMoney_Equal(t *testing.T) {
	tests := []struct {
		name string
		a, b Money
		want bool
	}{
		{"same", NewMoney(dec("10.00"), "PLN"), NewMoney(dec("10.00"), "PLN"), true},
		{"different amount", NewMoney(dec("10.00"), "PLN"), NewMoney(dec("20.00"), "PLN"), false},
		{"different currency", NewMoney(dec("10.00"), "PLN"), NewMoney(dec("10.00"), "EUR"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMoney_Comparisons(t *testing.T) {
	small := NewMoney(dec("5.00"), "PLN")
	big := NewMoney(dec("10.00"), "PLN")

	if !big.GreaterThan(small) {
		t.Error("expected 10 > 5")
	}
	if big.GreaterThan(big) {
		t.Error("expected 10 not > 10")
	}
	if !small.LessThanOrEqual(big) {
		t.Error("expected 5 <= 10")
	}
	if !big.LessThanOrEqual(big) {
		t.Error("expected 10 <= 10")
	}
}

func TestMoney_SignChecks(t *testing.T) {
	if !Zero("PLN").IsZero() {
		t.Error("Zero should be zero")
	}
	if !NewMoney(dec("-1"), "PLN").IsNegative() {
		t.Error("-1 should be negative")
	}
	if !NewMoney(dec("1"), "PLN").IsPositive() {
		t.Error("1 should be positive")
	}
	if NewMoney(dec("1"), "PLN").IsZero() {
		t.Error("1 should not be zero")
	}
}

func TestMoney_String(t *testing.T) {
	tests := []struct {
		name string
		m    Money
		want string
	}{
		{"two decimals", NewMoney(dec("10.50"), "PLN"), "10.50 PLN"},
		{"integer padded", NewMoney(dec("100"), "EUR"), "100.00 EUR"},
		{"many decimals truncated", NewMoney(dec("1.999"), "PLN"), "2.00 PLN"},
		{"zero", Zero("PLN"), "0.00 PLN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
