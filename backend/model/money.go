package model

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const DefaultCurrency = "PLN"

// Money is an immutable value object representing a monetary amount with currency.
// Fields are unexported to enforce immutability — use constructors and accessors.
type Money struct {
	amount   decimal.Decimal
	currency string
}

// NewMoney creates a Money value. If currency is empty, defaults to PLN.
func NewMoney(amount decimal.Decimal, currency string) Money {
	if currency == "" {
		currency = DefaultCurrency
	}
	return Money{amount: amount, currency: currency}
}

// NewMoneyFromString parses a string amount and returns a Money or an error.
func NewMoneyFromString(amount string, currency string) (Money, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		return Money{}, fmt.Errorf("invalid money amount %q: %w", amount, err)
	}
	return NewMoney(d, currency), nil
}

// Zero returns a zero Money value for the given currency.
func Zero(currency string) Money {
	if currency == "" {
		currency = DefaultCurrency
	}
	return Money{amount: decimal.Zero, currency: currency}
}

// Amount returns the decimal amount.
func (m Money) Amount() decimal.Decimal { return m.amount }

// Currency returns the currency code.
func (m Money) Currency() string { return m.currency }

// Add returns a new Money that is the sum of m and other.
// Panics if currencies differ.
func (m Money) Add(other Money) Money {
	m.mustMatchCurrency(other)
	return Money{amount: m.amount.Add(other.amount), currency: m.currency}
}

// Sub returns a new Money that is m minus other.
// Panics if currencies differ.
func (m Money) Sub(other Money) Money {
	m.mustMatchCurrency(other)
	return Money{amount: m.amount.Sub(other.amount), currency: m.currency}
}

// Equal returns true if both amount and currency match.
func (m Money) Equal(other Money) bool {
	return m.currency == other.currency && m.amount.Equal(other.amount)
}

// GreaterThan returns true if m is greater than other.
// Panics if currencies differ.
func (m Money) GreaterThan(other Money) bool {
	m.mustMatchCurrency(other)
	return m.amount.GreaterThan(other.amount)
}

// LessThanOrEqual returns true if m is less than or equal to other.
// Panics if currencies differ.
func (m Money) LessThanOrEqual(other Money) bool {
	m.mustMatchCurrency(other)
	return m.amount.LessThanOrEqual(other.amount)
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool { return m.amount.IsZero() }

// IsNegative returns true if the amount is negative.
func (m Money) IsNegative() bool { return m.amount.IsNegative() }

// IsPositive returns true if the amount is positive.
func (m Money) IsPositive() bool { return m.amount.IsPositive() }

// String returns the amount with 2 decimal places followed by the currency code.
func (m Money) String() string {
	return m.amount.StringFixed(2) + " " + m.currency
}

func (m Money) mustMatchCurrency(other Money) {
	if m.currency != other.currency {
		panic(fmt.Sprintf("currency mismatch: %s vs %s", m.currency, other.currency))
	}
}
