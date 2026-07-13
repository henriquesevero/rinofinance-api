package mongodb

import (
	"testing"

	"rinofinance-api/internal/domain/shared"
)

func TestDecimal128RoundTrip(t *testing.T) {
	cases := []string{"0", "1234.56", "0.10", "999999999.99", "-42.05"}

	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			original, err := shared.NewMoney(raw)
			if err != nil {
				t.Fatalf("unexpected error building money %q: %v", raw, err)
			}

			d, err := toDecimal128(original)
			if err != nil {
				t.Fatalf("toDecimal128 error: %v", err)
			}

			roundTripped, err := fromDecimal128(d)
			if err != nil {
				t.Fatalf("fromDecimal128 error: %v", err)
			}

			if !original.Decimal().Equal(roundTripped.Decimal()) {
				t.Errorf("round trip mismatch: got %s, want %s", roundTripped, original)
			}
		})
	}
}
