package db

import (
	"database/sql"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	const pragmas = "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"

	dsn := path + pragmas

	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	db := &DB{DB: sqlDB}

	if err := db.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	return db, nil
}

func (db *DB) migrate() error {
	schema := `
		create table if not exists rates (
			date        date      not null,
			base        text      not null,
			target      text      not null,
			rate        real      not null,
			source      text      not null,
			calculated  integer   not null default 0,
			fetched_at  timestamp not null,
			primary key (date, base, target)
		);

		create index if not exists idx_rates_date      on rates(date);
		create index if not exists idx_rates_base_date on rates(base, date);
		create index if not exists idx_rates_source    on rates(source);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("creating schema: %w", err)
	}

	return nil
}
