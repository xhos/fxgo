package common

import (
	"fmt"
	"time"

	"github.com/xhos/fxgo/internal/models"
)

func ValidateRates(rates []models.Rate) error {
	if len(rates) == 0 {
		return fmt.Errorf("no rates returned")
	}

	for i, rate := range rates {
		if err := validateRate(i, rate); err != nil {
			return err
		}
	}

	return nil
}

func validateRate(index int, rate models.Rate) error {
	invalidValue := rate.Value <= 0
	if invalidValue {
		return fmt.Errorf("rate[%d]: invalid value %.4f", index, rate.Value)
	}

	emptyBase := rate.Base == ""
	if emptyBase {
		return fmt.Errorf("rate[%d]: empty base currency", index)
	}

	emptyTarget := rate.Target == ""
	if emptyTarget {
		return fmt.Errorf("rate[%d]: empty target currency", index)
	}

	selfConversion := rate.Base == rate.Target
	if selfConversion {
		return fmt.Errorf("rate[%d]: base and target are the same (%s)", index, rate.Base)
	}

	zeroDate := rate.Date.IsZero()
	if zeroDate {
		return fmt.Errorf("rate[%d]: zero date", index)
	}

	maxAllowedDate := time.Now().Add(24 * time.Hour)
	futureDate := rate.Date.After(maxAllowedDate)
	if futureDate {
		return fmt.Errorf("rate[%d]: future date %s", index, rate.Date)
	}

	emptySource := rate.Source == ""
	if emptySource {
		return fmt.Errorf("rate[%d]: empty source", index)
	}

	return nil
}
