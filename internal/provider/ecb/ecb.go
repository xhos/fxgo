package ecb

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
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
		baseURL: "https://data-api.ecb.europa.eu",
		client:  common.NewHTTPClient(30 * time.Second),
	}
}

func (p *Provider) Name() string {
	return "ECB"
}

func (p *Provider) FetchRates(ctx context.Context, req models.RateRequest) ([]models.Rate, error) {
	isDirectEUR := (req.Base == "EUR")
	if isDirectEUR {
		return p.fetchDirectRates(ctx, req)
	}

	return p.fetchCrossRates(ctx, req)
}

func (p *Provider) fetchDirectRates(ctx context.Context, req models.RateRequest) ([]models.Rate, error) {
	url := p.buildURL(req.Targets, req.Date)

	body, err := p.client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching from ecb: %w", err)
	}

	rates, err := p.parseCSV(body, "EUR", req.Date)
	if err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if err := common.ValidateRates(rates); err != nil {
		return nil, fmt.Errorf("validating rates: %w", err)
	}

	return rates, nil
}

func (p *Provider) fetchCrossRates(ctx context.Context, req models.RateRequest) ([]models.Rate, error) {
	// fetch both base and targets to calculate cross-rates via EUR
	allCurrencies := append([]string{req.Base}, req.Targets...)
	url := p.buildURL(allCurrencies, req.Date)

	body, err := p.client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching from ecb: %w", err)
	}

	eurRates, err := p.parseCSV(body, "EUR", req.Date)
	if err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if err := common.ValidateRates(eurRates); err != nil {
		return nil, fmt.Errorf("validating eur rates: %w", err)
	}

	rates, err := common.CalculateCrossRates(eurRates, req.Base, req.Targets)
	if err != nil {
		return nil, fmt.Errorf("calculating cross rates: %w", err)
	}

	return rates, nil
}

func (p *Provider) buildURL(currencies []string, date time.Time) string {
	currencyList := strings.Join(currencies, "+")
	seriesKey := fmt.Sprintf("D.%s.EUR.SP00.A", currencyList)

	url := fmt.Sprintf("%s/service/data/EXR/%s?format=csvdata", p.baseURL, seriesKey)

	hasSpecificDate := !date.IsZero()
	if hasSpecificDate {
		dateStr := date.Format("2006-01-02")
		url += fmt.Sprintf("&startPeriod=%s&endPeriod=%s", dateStr, dateStr)
	} else {
		url += "&lastNObservations=1"
	}

	return url
}

func (p *Provider) parseCSV(data []byte, base string, requestDate time.Time) ([]models.Rate, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading csv: %w", err)
	}

	noData := (len(records) < 2)
	if noData {
		return nil, fmt.Errorf("no data in response")
	}

	header := records[0]
	currencyIdx := p.findColumnIndex(header, "CURRENCY")
	dateIdx := p.findColumnIndex(header, "TIME_PERIOD")
	valueIdx := p.findColumnIndex(header, "OBS_VALUE")

	missingColumns := (currencyIdx == -1 || dateIdx == -1 || valueIdx == -1)
	if missingColumns {
		return nil, fmt.Errorf("missing required csv columns")
	}

	var rates []models.Rate
	now := time.Now()

	for _, record := range records[1:] {
		rate, skip := p.parseCSVRecord(record, currencyIdx, dateIdx, valueIdx, base, now, requestDate)
		if skip {
			continue
		}
		rates = append(rates, rate)
	}

	return rates, nil
}

func (p *Provider) findColumnIndex(header []string, columnName string) int {
	for i, col := range header {
		if col == columnName {
			return i
		}
	}
	return -1
}

func (p *Provider) parseCSVRecord(record []string, currencyIdx, dateIdx, valueIdx int, base string, fetchedAt time.Time, requestDate time.Time) (models.Rate, bool) {
	insufficientColumns := (len(record) <= currencyIdx || len(record) <= dateIdx || len(record) <= valueIdx)
	if insufficientColumns {
		return models.Rate{}, true
	}

	currency := record[currencyIdx]
	dateStr := record[dateIdx]
	valueStr := record[valueIdx]

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return models.Rate{}, true
	}

	hasSpecificDate := !requestDate.IsZero()
	wrongDate := hasSpecificDate && !date.Equal(requestDate.Truncate(24*time.Hour))
	if wrongDate {
		return models.Rate{}, true
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return models.Rate{}, true
	}

	return models.Rate{
		Base:       base,
		Target:     currency,
		Value:      value,
		Date:       date,
		Source:     "ECB",
		Fetched:    fetchedAt,
		Calculated: false,
	}, false
}

// TODO: a more dynamic approach could be implemented
func (p *Provider) SupportedCurrencies() []string {
	// actively updated currencies as of 2025
	return []string{
		"AUD", "BGN", "BRL", "CAD", "CHF", "CNY", "CZK",
		"DKK", "GBP", "HKD", "HUF", "IDR", "ILS", "INR",
		"ISK", "JPY", "KRW", "MXN", "MYR", "NOK", "NZD",
		"PHP", "PLN", "RON", "SEK", "SGD", "THB", "TRY",
		"USD", "ZAR",
	}
}
