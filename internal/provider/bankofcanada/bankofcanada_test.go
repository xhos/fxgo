package bankofcanada

import (
	"context"
	"testing"

	"github.com/xhos/fxgo/internal/models"
)

func TestFetchRates(t *testing.T) {
	p := New()
	ctx := context.Background()

	t.Run("direct CAD rates with inversion", func(t *testing.T) {
		rates, err := p.FetchRates(ctx, models.RateRequest{
			Base:    "CAD",
			Targets: []string{"USD", "EUR"},
		})

		if err != nil {
			t.Fatal(err)
		}

		if len(rates) != 2 {
			t.Fatalf("got %d rates, want 2", len(rates))
		}

		for _, r := range rates {
			positiveValue := r.Value > 0
			correctBase := r.Base == "CAD"
			correctSource := r.Source == "BankOfCanada"
			isDirect := !r.Calculated

			if !positiveValue || !correctBase || !correctSource || !isDirect {
				t.Errorf("invalid rate: %+v", r)
			}
		}
	})

	t.Run("cross rates via CAD", func(t *testing.T) {
		rates, err := p.FetchRates(ctx, models.RateRequest{
			Base:    "USD",
			Targets: []string{"EUR", "JPY"},
		})

		if err != nil {
			t.Fatal(err)
		}

		if len(rates) != 2 {
			t.Fatalf("got %d rates, want 2", len(rates))
		}

		for _, r := range rates {
			positiveValue := r.Value > 0
			correctBase := r.Base == "USD"
			isCalculated := r.Calculated

			if !positiveValue || !correctBase || !isCalculated {
				t.Errorf("invalid cross-rate: %+v", r)
			}
		}
	})
}

func TestExtractCurrency(t *testing.T) {
	p := New()

	cases := map[string]string{
		"FXUSDCAD": "USD",
		"FXEURCAD": "EUR",
		"INVALID":  "",
		"FXUSD":    "",
	}

	for input, want := range cases {
		got := p.extractCurrency(input)
		if got != want {
			t.Errorf("extractCurrency(%s) = %s, want %s", input, got, want)
		}
	}
}
