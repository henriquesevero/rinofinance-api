package card

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

func mustMoney(t *testing.T, v string) shared.Money {
	t.Helper()
	m, err := shared.NewMoney(v)
	if err != nil {
		t.Fatalf("unexpected error building money %q: %v", v, err)
	}
	return m
}

func TestInstallmentPurchase_RemainingInstallments(t *testing.T) {
	first := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	p, err := NewInstallmentPurchase(uuid.New(), "Notebook", mustMoney(t, "500.00"), 10, first)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := []struct {
		name             string
		reference        time.Time
		wantRemaining    int
		wantActive       bool
		wantChargeIsZero bool
	}{
		{"month before first installment", time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC), 10, false, true},
		{"first installment month", time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC), 10, true, false},
		{"middle installment month", time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), 5, true, false},
		{"last installment month", time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC), 1, true, false},
		{"month after last installment", time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC), 0, false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := p.RemainingInstallments(tc.reference); got != tc.wantRemaining {
				t.Errorf("RemainingInstallments = %d, want %d", got, tc.wantRemaining)
			}
			if got := p.IsActiveOn(tc.reference); got != tc.wantActive {
				t.Errorf("IsActiveOn = %v, want %v", got, tc.wantActive)
			}
			gotZero := p.MonthlyChargeAmount(tc.reference).IsZero()
			if gotZero != tc.wantChargeIsZero {
				t.Errorf("MonthlyChargeAmount zero = %v, want %v", gotZero, tc.wantChargeIsZero)
			}
		})
	}
}

func TestInstallmentPurchase_RemainingTotal(t *testing.T) {
	first := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	p, err := NewInstallmentPurchase(uuid.New(), "TV", mustMoney(t, "200.00"), 5, first)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reference := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	got := p.RemainingTotal(reference)
	want := mustMoney(t, "600.00")
	if !got.Decimal().Equal(want.Decimal()) {
		t.Errorf("RemainingTotal = %s, want %s", got, want)
	}
}

func TestMonthlyTotal_SumsPurchasesAndSubscriptions(t *testing.T) {
	cardID := uuid.New()
	reference := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)

	active, err := NewInstallmentPurchase(cardID, "Celular", mustMoney(t, "300.00"), 12, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	finished, err := NewInstallmentPurchase(cardID, "Fone", mustMoney(t, "100.00"), 3, time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sub, err := NewSubscription(cardID, "Streaming", mustMoney(t, "39.90"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	total := MonthlyTotal(reference, []*InstallmentPurchase{active, finished}, []*Subscription{sub})
	want := mustMoney(t, "339.90")
	if !total.Decimal().Equal(want.Decimal()) {
		t.Errorf("MonthlyTotal = %s, want %s (finished purchase must not be counted)", total, want)
	}
}
