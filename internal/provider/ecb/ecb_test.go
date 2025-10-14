package ecb

import (
	"context"
	"testing"

	"github.com/xhos/fxgo/internal/models"
)

func TestFetchRates(t *testing.T) {
	p := New()
	ctx := context.Background()

	t.Run("direct EUR rates", func(t *testing.T) {
		rates, err := p.FetchRates(ctx, models.RateRequest{
			Base:    "EUR",
			Targets: []string{"USD", "GBP", "JPY"},
		})

		if err != nil {
			t.Fatal(err)
		}

		if len(rates) != 3 {
			t.Fatalf("got %d rates, want 3", len(rates))
		}

		for _, r := range rates {
			positiveValue := r.Value > 0
			correctBase := r.Base == "EUR"
			correctSource := r.Source == "ECB"
			isDirect := !r.Calculated

			if !positiveValue || !correctBase || !correctSource || !isDirect {
				t.Errorf("invalid rate: %+v", r)
			}
		}
	})

	t.Run("cross rates via EUR", func(t *testing.T) {
		rates, err := p.FetchRates(ctx, models.RateRequest{
			Base:    "USD",
			Targets: []string{"JPY", "GBP"},
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
