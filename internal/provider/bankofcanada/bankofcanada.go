package bankofcanada

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xhos/fxgo/internal/models"
	"github.com/xhos/fxgo/internal/provider/common"
)

type Provider struct {
	baseURL string
	client  *common.HTTPClient
}

func New() *Provider {
	return &Provider{
		baseURL: "https://www.bankofcanada.ca/valet",
		client:  common.NewHTTPClient(30 * time.Second),
	}
}

func (p *Provider) Name() string {
	return "BankOfCanada"
}

func (p *Provider) FetchRates(ctx context.Context, req models.RateRequest) ([]models.Rate, error) {
	isDirectCAD := (req.Base == "CAD")
	if isDirectCAD {
		return p.fetchDirectRates(ctx, req)
	}

	return p.fetchCrossRates(ctx, req)
}

func (p *Provider) fetchDirectRates(ctx context.Context, req models.RateRequest) ([]models.Rate, error) {
	seriesNames := p.buildSeriesNames(req.Targets)
	url := p.buildURL(seriesNames, req.Date)

	body, err := p.client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching from bank of canada: %w", err)
	}

	var data response
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	rates := p.parseResponse(data, "CAD", req.Date)
	if err := common.ValidateRates(rates); err != nil {
		return nil, fmt.Errorf("validating rates: %w", err)
	}

	return rates, nil
}

func (p *Provider) fetchCrossRates(ctx context.Context, req models.RateRequest) ([]models.Rate, error) {

	// fetch both base and targets to calculate cross-rates via CAD
	allCurrencies := append([]string{req.Base}, req.Targets...)
	seriesNames := p.buildSeriesNames(allCurrencies)
	url := p.buildURL(seriesNames, req.Date)

	body, err := p.client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching from bank of canada: %w", err)
	}

	var data response
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	cadRates := p.parseResponse(data, "CAD", req.Date)
	if err := common.ValidateRates(cadRates); err != nil {
		return nil, fmt.Errorf("validating cad rates: %w", err)
	}

	rates, err := common.CalculateCrossRates(cadRates, req.Base, req.Targets)
	if err != nil {
		return nil, fmt.Errorf("calculating cross rates: %w", err)
	}

	return rates, nil
}

// buildSeriesNames converts currency codes to BoC series names (e.g., "USD" -> "FXUSDCAD")
func (p *Provider) buildSeriesNames(currencies []string) []string {
	var series []string
	for _, currency := range currencies {
		seriesName := fmt.Sprintf("FX%sCAD", currency)
		series = append(series, seriesName)
	}
	return series
}

func (p *Provider) buildURL(seriesNames []string, date time.Time) string {
	series := strings.Join(seriesNames, ",")
	url := fmt.Sprintf("%s/observations/%s/json", p.baseURL, series)

	hasSpecificDate := !date.IsZero()
	if hasSpecificDate {
		dateStr := date.Format("2006-01-02")
		url += fmt.Sprintf("?start_date=%s&end_date=%s", dateStr, dateStr)
	} else {
		url += "?recent=1"
	}

	return url
}

// parseResponse extracts rates from BoC API response
// BoC returns foreign-to-CAD rates, so we invert them for CAD-based queries
func (p *Provider) parseResponse(data response, base string, requestDate time.Time) []models.Rate {
	var rates []models.Rate
	now := time.Now()

	for _, obs := range data.Observations {
		date, skip := p.parseDate(obs, requestDate)
		if skip {
			continue
		}

		for key, val := range obs {
			isDateField := (key == "d")
			if isDateField {
				continue
			}

			rate, skip := p.parseRate(key, val, base, date, now)
			if skip {
				continue
			}

			rates = append(rates, rate)
		}
	}

	return rates
}

func (p *Provider) parseDate(obs map[string]any, requestDate time.Time) (time.Time, bool) {
	dateStr, ok := obs["d"].(string)
	if !ok {
		return time.Time{}, true
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, true
	}

	hasSpecificDate := !requestDate.IsZero()
	wrongDate := hasSpecificDate && !isSameDay(date, requestDate)
	if wrongDate {
		return time.Time{}, true
	}

	return date, false
}

func (p *Provider) parseRate(key string, val any, base string, date time.Time, fetchedAt time.Time) (models.Rate, bool) {
	currency := p.extractCurrency(key)
	if currency == "" {
		return models.Rate{}, true
	}

	valueMap, ok := val.(map[string]any)
	if !ok {
		return models.Rate{}, true
	}

	valueStr, ok := valueMap["v"].(string)
	if !ok {
		return models.Rate{}, true
	}

	var value float64
	if _, err := fmt.Sscanf(valueStr, "%f", &value); err != nil {
		return models.Rate{}, true
	}

	invalidValue := (value <= 0)
	if invalidValue {
		return models.Rate{}, true
	}

	invertedValue := 1.0 / value

	return models.Rate{
		Base:       base,
		Target:     currency,
		Value:      invertedValue,
		Date:       date,
		Source:     "BankOfCanada",
		Fetched:    fetchedAt,
		Calculated: false,
	}, false
}

// extractCurrency converts BoC series name to currency code (e.g., "FXUSDCAD" -> "USD")
func (p *Provider) extractCurrency(seriesName string) string {
	hasPrefix := strings.HasPrefix(seriesName, "FX")
	hasSuffix := strings.HasSuffix(seriesName, "CAD")

	validFormat := hasPrefix && hasSuffix
	if !validFormat {
		return ""
	}

	currency := strings.TrimPrefix(seriesName, "FX")
	currency = strings.TrimSuffix(currency, "CAD")
	return currency
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// TODO: a more dynamic approach could be implemented
func (p *Provider) SupportedCurrencies() []string {
	// actively updated currencies as of 2025
	return []string{
		"AUD", "BRL", "CNY", "EUR", "HKD", "INR", "IDR",
		"JPY", "MXN", "NZD", "NOK", "PEN", "RUB", "SAR",
		"SGD", "ZAR", "KRW", "SEK", "CHF", "TWD", "TRY",
		"GBP", "USD",
	}
}
