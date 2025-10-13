package db

import (
	"context"
	"fmt"
	"time"

	"github.com/xhos/fxgo/internal/models"
)

func (db *DB) GetRatesBetween(ctx context.Context, startDate, endDate time.Time, base string, targets []string) ([]models.Rate, error) {
	if len(targets) == 0 {
		return []models.Rate{}, nil
	}

	queryTemplate := `
		select date, base, target, rate, source, calculated, fetched_at
		from   rates
		where  base = ? and date >= ? and date <= ? and target in (%s)
		order by date asc, target asc
	`

	placeholders, args := db.buildInClause([]any{base, startDate, endDate}, targets)
	query := fmt.Sprintf(queryTemplate, placeholders)

	rates, err := db.scanMultipleRates(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying rates between dates: %w", err)
	}

	return rates, nil
}

func (db *DB) GetAvailableCurrencies(ctx context.Context) ([]string, error) {
	query := `
		select distinct target
		from   rates
		where  calculated = 0
		order by target
	`

	currencies, err := db.scanStringList(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying available currencies: %w", err)
	}

	return currencies, nil
}

func (db *DB) GetAvailableBases(ctx context.Context) ([]string, error) {
	query := `
		select distinct base
		from   rates
		where  calculated = 0
		order by base
	`

	bases, err := db.scanStringList(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying available bases: %w", err)
	}

	return bases, nil
}

func (db *DB) scanStringList(ctx context.Context, query string, args ...any) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		result = append(result, value)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (db *DB) GetDateRange(ctx context.Context) (time.Time, time.Time, error) {
	query := `
		select min(date), max(date)
		from   rates
	`

	var minDate, maxDate time.Time
	err := db.QueryRowContext(ctx, query).Scan(&minDate, &maxDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("querying date range: %w", err)
	}

	return minDate, maxDate, nil
}
