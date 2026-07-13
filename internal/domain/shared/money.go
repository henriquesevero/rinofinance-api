// Package shared contains value objects and building blocks shared across
// every bounded context in the domain layer.
package shared

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Money represents a monetary amount. It wraps decimal.Decimal instead of
// float64 so currency values never suffer from binary floating point
// rounding errors when summed across incomes, expenses, card installments
// and investment balances.
type Money struct {
	amount decimal.Decimal
}

// Zero is the additive identity for Money.
var Zero = Money{amount: decimal.Zero}

// NewMoney builds a Money from a decimal string (e.g. "1234.56"). It is the
// preferred constructor when parsing values coming from persistence or the
// HTTP layer, since it avoids the precision loss of float64.
func NewMoney(value string) (Money, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return Money{}, fmt.Errorf("%w: valor monetário inválido %q", ErrInvalidAmount, value)
	}
	return Money{amount: d}, nil
}

// NewMoneyFromFloat builds a Money from a float64. Prefer NewMoney when the
// original representation is a string, this constructor exists mainly for
// convenience at the edges (tests, CLI seeding).
func NewMoneyFromFloat(value float64) Money {
	return Money{amount: decimal.NewFromFloat(value)}
}

// NewMoneyFromDecimal builds a Money directly from a decimal.Decimal. It
// exists for the Postgres adapter, which scans NUMERIC columns straight
// into decimal.Decimal and needs to wrap them without a lossy string
// round-trip.
func NewMoneyFromDecimal(value decimal.Decimal) Money {
	return Money{amount: value}
}

// IsNegative reports whether the amount is strictly less than zero.
func (m Money) IsNegative() bool {
	return m.amount.IsNegative()
}

// IsZero reports whether the amount is exactly zero.
func (m Money) IsZero() bool {
	return m.amount.IsZero()
}

// Add returns the sum of two Money values.
func (m Money) Add(other Money) Money {
	return Money{amount: m.amount.Add(other.amount)}
}

// Sub returns the difference between two Money values (e.g. income minus
// expenses when computing the dashboard's net balance). The result may be
// negative.
func (m Money) Sub(other Money) Money {
	return Money{amount: m.amount.Sub(other.amount)}
}

// MulInt returns the amount multiplied by an integer factor. Used to
// compute totals such as installment amount × remaining installments.
func (m Money) MulInt(factor int) Money {
	return Money{amount: m.amount.Mul(decimal.NewFromInt(int64(factor)))}
}

// Decimal exposes the underlying decimal.Decimal, needed by adapters
// (Postgres driver, JSON encoding) that must serialize the exact value.
func (m Money) Decimal() decimal.Decimal {
	return m.amount
}

// String renders the amount with two decimal places (e.g. "1234.56").
func (m Money) String() string {
	return m.amount.StringFixed(2)
}

// MarshalJSON encodes Money as an unquoted JSON number, matching how the
// frontend expects amounts (a plain decimal number it can do arithmetic on,
// not a quoted string). decimal.MarshalJSON defaults to a quoted string, so
// we emit the decimal's plain representation directly instead.
func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(m.amount.String()), nil
}

// UnmarshalJSON decodes Money from a JSON number or numeric string.
func (m *Money) UnmarshalJSON(data []byte) error {
	var d decimal.Decimal
	if err := d.UnmarshalJSON(data); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAmount, err)
	}
	m.amount = d
	return nil
}
