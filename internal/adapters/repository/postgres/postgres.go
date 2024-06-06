package postgres

import (
	"database/sql"
	"fmt"

	"github.com/4aykovski/url_shortener/internal/config"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func New(cfg config.Postgres) (*Postgres, error) {
	const op = "Postgres.Postgres.New"

	db, err := sql.Open("postgres", cfg.DSNTemplate)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Postgres{db: db}, nil
}
