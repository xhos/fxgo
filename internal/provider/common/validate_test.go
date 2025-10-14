package common

import (
	"testing"
	"time"

	"github.com/xhos/fxgo/internal/models"
)

func TestValidateRates(t *testing.T) {
	now := time.Now()

	valid := models.Rate{
		Base:   "EUR",
		Target: "USD",
		Value:  1.15,
		Date:   now,
		Source: "ECB",
	}

	shouldPass := []models.Rate{valid}
	shouldFail := map[string][]models.Rate{
		"empty":           {},
		"negative value":  {{Base: "EUR", Target: "USD", Value: -1, Date: now, Source: "ECB"}},
		"zero value":      {{Base: "EUR", Target: "USD", Value: 0, Date: now, Source: "ECB"}},
		"self conversion": {{Base: "EUR", Target: "EUR", Value: 1, Date: now, Source: "ECB"}},
		"empty base":      {{Base: "", Target: "USD", Value: 1, Date: now, Source: "ECB"}},
		"future date":     {{Base: "EUR", Target: "USD", Value: 1, Date: now.Add(48 * time.Hour), Source: "ECB"}},
	}

	if err := ValidateRates(shouldPass); err != nil {
		t.Errorf("valid rate rejected: %v", err)
	}

	for name, rates := range shouldFail {
		if err := ValidateRates(rates); err == nil {
			t.Errorf("%s: expected error, got none", name)
		}
	}
}
