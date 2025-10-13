package provider

import (
	"context"

	"github.com/xhos/fxgo/internal/models"
)

type Provider interface {
	Name() string
	FetchRates(ctx context.Context, req models.RateRequest) ([]models.Rate, error)
	SupportedCurrencies() []string
}
