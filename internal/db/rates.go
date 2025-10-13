package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/xhos/fxgo/internal/models"
)

func (db *DB) InsertRate(ctx context.Context, rate models.Rate) error {
	query := `
		insert into rates (date, base, target, rate, source, calculated, fetched_at)
		values (?, ?, ?, ?, ?, ?, ?)
		on conflict (date, base, target) do update set
			rate       = excluded.rate,
			source     = excluded.source,
			calculated = excluded.calculated,
			fetched_at = excluded.fetched_at
	`

	_, err := db.ExecContext(ctx, query,
		rate.Date,
		rate.Base,
		rate.Target,
		rate.Value,
		rate.Source,
		rate.Calculated,
		rate.Fetched,
	)

	if err != nil {
		return fmt.Errorf("inserting rate: %w", err)
	}

	return nil
}

func (db *DB) InsertRates(ctx context.Context, rates []models.Rate) error {
	if len(rates) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	upsertQuery := `
		insert into rates (date, base, target, rate, source, calculated, fetched_at)
		values (?, ?, ?, ?, ?, ?, ?)
		on conflict (date, base, target) do update set
			rate       = excluded.rate,
			source     = excluded.source,
			calculated = excluded.calculated,
			fetched_at = excluded.fetched_at
	`
	stmt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, rate := range rates {
		_, err := stmt.ExecContext(ctx,
			rate.Date,
			rate.Base,
			rate.Target,
			rate.Value,
			rate.Source,
			rate.Calculated,
			rate.Fetched,
		)
		if err != nil {
			return fmt.Errorf("inserting rate: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (db *DB) GetRate(ctx context.Context, date time.Time, base, target string) (*models.Rate, error) {
	query := `
		select date, base, target, rate, source, calculated, fetched_at
		from   rates
		where  date = ? and base = ? and target = ?
	`

	rate, err := db.scanSingleRate(ctx, query, date, base, target)
	if err != nil {
		return nil, fmt.Errorf("querying rate: %w", err)
	}

	return rate, nil
}

func (db *DB) GetLatestRate(ctx context.Context, base, target string) (*models.Rate, error) {
	query := `
		select date, base, target, rate, source, calculated, fetched_at
		from   rates
		where  base = ? and target = ?
		order by date desc
		limit  1
	`

	rate, err := db.scanSingleRate(ctx, query, base, target)
	if err != nil {
		return nil, fmt.Errorf("querying latest rate: %w", err)
	}

	return rate, nil
}

func (db *DB) GetRatesForDate(ctx context.Context, date time.Time, base string, targets []string) ([]models.Rate, error) {
	if len(targets) == 0 {
		return []models.Rate{}, nil
	}

	queryTemplate := `
		select date, base, target, rate, source, calculated, fetched_at
		from   rates
		where  date = ? and base = ? and target in (%s)
	`

	placeholders, args := db.buildInClause([]any{date, base}, targets)
	query := fmt.Sprintf(queryTemplate, placeholders)

	rates, err := db.scanMultipleRates(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying rates: %w", err)
	}

	return rates, nil
}

func (db *DB) GetLatestRates(ctx context.Context, base string, targets []string) ([]models.Rate, error) {
	if len(targets) == 0 {
		return []models.Rate{}, nil
	}

	// join with subquery to get most recent date for each target
	queryTemplate := `
		select r1.date, r1.base, r1.target, r1.rate, r1.source, r1.calculated, r1.fetched_at
		from   rates r1
		inner join (
			select target, max(date) as max_date
			from   rates
			where  base = ? and target in (%s)
			group by target
		) r2 on r1.target = r2.target and r1.date = r2.max_date
		where  r1.base = ?
	`

	placeholders, args := db.buildInClause([]any{base}, targets)
	args = append(args, base)
	query := fmt.Sprintf(queryTemplate, placeholders)

	rates, err := db.scanMultipleRates(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying latest rates: %w", err)
	}

	return rates, nil
}

func (db *DB) GetNearestDate(ctx context.Context, date time.Time) (time.Time, error) {
	query := `
		select date
		from   rates
		where  date <= ?
		order by date desc
		limit  1
	`

	var nearestDate time.Time
	err := db.QueryRowContext(ctx, query, date).Scan(&nearestDate)

	notFound := (err == sql.ErrNoRows)
	if notFound {
		return time.Time{}, fmt.Errorf("no rates found before %s", date)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("querying nearest date: %w", err)
	}

	return nearestDate, nil
}

func (db *DB) scanSingleRate(ctx context.Context, query string, args ...any) (*models.Rate, error) {
	var rate models.Rate
	var calculatedInt int

	err := db.QueryRowContext(ctx, query, args...).Scan(
		&rate.Date,
		&rate.Base,
		&rate.Target,
		&rate.Value,
		&rate.Source,
		&calculatedInt,
		&rate.Fetched,
	)

	notFound := (err == sql.ErrNoRows)
	if notFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rate.Calculated = (calculatedInt != 0)

	return &rate, nil
}

func (db *DB) scanMultipleRates(ctx context.Context, query string, args ...any) ([]models.Rate, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []models.Rate
	for rows.Next() {
		var rate models.Rate
		var calculatedInt int

		err := rows.Scan(
			&rate.Date,
			&rate.Base,
			&rate.Target,
			&rate.Value,
			&rate.Source,
			&calculatedInt,
			&rate.Fetched,
		)
		if err != nil {
			return nil, err
		}

		rate.Calculated = (calculatedInt != 0)
		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rates, nil
}

func (db *DB) buildInClause(prefixArgs []any, inValues []string) (string, []any) {
	args := make([]any, len(prefixArgs))
	copy(args, prefixArgs)

	placeholders := ""
	for i, val := range inValues {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
		args = append(args, val)
	}

	return placeholders, args
}
