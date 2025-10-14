package common

import (
	"fmt"
	"time"

	"github.com/xhos/fxgo/internal/models"
)

// CalculateCrossRates converts rates from one base currency to another
// via cross-rate calculation: if X/USD = 1.1 and X/JPY = 130, then USD/JPY = 130/1.1
func CalculateCrossRates(sourceRates []models.Rate, base string, targets []string) ([]models.Rate, error) {
	baseRate := findRate(sourceRates, base)
	if baseRate == nil {
		return nil, fmt.Errorf("base currency %s not found", base)
	}

	var rates []models.Rate
	for _, target := range targets {
		targetRate := findRate(sourceRates, target)
		if targetRate == nil {
			continue
		}

		// target/base = (source/target) / (source/base)
		crossRate := targetRate.Value / baseRate.Value

		rates = append(rates, models.Rate{
			Base:       base,
			Target:     target,
			Value:      crossRate,
			Date:       targetRate.Date,
			Source:     targetRate.Source,
			Fetched:    time.Now(),
			Calculated: true,
		})
	}

	noTargetsFound := (len(rates) == 0)
	if noTargetsFound {
		return nil, fmt.Errorf("no target currencies found")
	}

	return rates, nil
}

func findRate(rates []models.Rate, currency string) *models.Rate {
	for i := range rates {
		isMatch := (rates[i].Target == currency)
		if isMatch {
			return &rates[i]
		}
	}
	return nil
}
