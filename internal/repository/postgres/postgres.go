package postgres

import (
	"database/sql"
	"fmt"

	"github.com/4aykovski/learning/golang/rest/internal/config"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func New(cfg config.Postgres) (*Postgres, error) {
	const op = "postgres.postgres.New"

	db, err := sql.Open("postgres", cfg.DSNTemplate)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = databasePrepare(db); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Postgres{db: db}, nil
}

func databasePrepare(db *sql.DB) error {
	const op = "postgres.postgres.databasePrepare"

	stmt1, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
	    id SERIAL PRIMARY KEY,
	    alias TEXT NOT NULL UNIQUE,
	    url TEXT NOT NULL
	);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt1.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt2, err := db.Prepare(`
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt2.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
