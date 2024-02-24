package Postgres

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
	const op = "Postgres.Postgres.New"

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
	const op = "Postgres.Postgres.databasePrepare"

	err := initUrlTable(db)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = initUsersTable(db)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = initRefreshSessionsTable(db)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func initUrlTable(db *sql.DB) error {
	const op = "Postgres.Postgres.initUrlTable"

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

func initUsersTable(db *sql.DB) error {
	const op = "Postgres.Postgres.initUsersTable"

	stmt1, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS "users"(
		"id" SERIAL PRIMARY KEY,
		"login" VARCHAR(128) NOT NULL UNIQUE,
		"password" VARCHAR(60) NOT NULL
	);
	`)
	if err != nil {
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	_, err = stmt1.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt2, err := db.Prepare(`
	CREATE INDEX IF NOT EXISTS idx_login ON "users"(login);
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt2.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func initRefreshSessionsTable(db *sql.DB) error {
	const op = "Postgres.Postgres.initRefreshSessionsTable"

	stmt1, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS "refresh_sessions"(
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) NOT NULL,
		refresh_token VARCHAR(128) NOT NULL UNIQUE,
		expires_in TIMESTAMP NOT NULL
	);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt1.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt2, err := db.Prepare(`
	CREATE INDEX IF NOT EXISTS idx_user_id ON "refresh_sessions"(refresh_token);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt2.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
