package shared

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Money struct {
	amount decimal.Decimal
}

var Zero = Money{amount: decimal.Zero}

func NewMoney(value string) (Money, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return Money{}, fmt.Errorf("%w: valor monetário inválido %q", ErrInvalidAmount, value)
	}
	return Money{amount: d}, nil
}

func NewMoneyFromFloat(value float64) Money {
	return Money{amount: decimal.NewFromFloat(value)}
}

func NewMoneyFromDecimal(value decimal.Decimal) Money {
	return Money{amount: value}
}

func (m Money) IsNegative() bool {
	return m.amount.IsNegative()
}

func (m Money) IsZero() bool {
	return m.amount.IsZero()
}

func (m Money) Add(other Money) Money {
	return Money{amount: m.amount.Add(other.amount)}
}

func (m Money) Sub(other Money) Money {
	return Money{amount: m.amount.Sub(other.amount)}
}

func (m Money) MulInt(factor int) Money {
	return Money{amount: m.amount.Mul(decimal.NewFromInt(int64(factor)))}
}

func (m Money) Decimal() decimal.Decimal {
	return m.amount
}

func (m Money) String() string {
	return m.amount.StringFixed(2)
}

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(m.amount.String()), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	var d decimal.Decimal
	if err := d.UnmarshalJSON(data); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAmount, err)
	}
	m.amount = d
	return nil
}
