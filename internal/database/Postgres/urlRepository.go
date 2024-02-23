package Postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/4aykovski/learning/golang/rest/internal/database"
	"github.com/lib/pq"
)

type UrlRepositoryPostgres struct {
	postgres *Postgres
}

func NewUserRepository(pq *Postgres) *UrlRepositoryPostgres {
	return &UrlRepositoryPostgres{postgres: pq}
}

func (repo *UrlRepositoryPostgres) SaveURL(urlToSave string, alias string) error {
	stmt, err := repo.postgres.db.Prepare("INSERT INTO url(url, alias) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("%w: %w", database.ErrCantSavePage, err)
	} // 23505

	_, err = stmt.Exec(urlToSave, alias)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return fmt.Errorf("%w: %w", database.ErrCantSavePage, database.ErrUrlExists)
			}
		}
		return fmt.Errorf("%w: %w", database.ErrCantSavePage, err)
	}

	return nil
}

func (repo *UrlRepositoryPostgres) GetURL(alias string) (string, error) {
	stmt, err := repo.postgres.db.Prepare("SELECT url FROM url WHERE alias=$1")
	if err != nil {
		return "", fmt.Errorf("%w: %w", database.ErrCantGetUrl, err)
	}

	var resultUrl string
	err = stmt.QueryRow(alias).Scan(&resultUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", database.ErrURLNotFound
		}

		return "", fmt.Errorf("%w: %w", database.ErrCantGetUrl, err)
	}

	return resultUrl, nil
}

func (repo *UrlRepositoryPostgres) DeleteURL(alias string) error {
	stmt, err := repo.postgres.db.Prepare("DELETE FROM url WHERE alias = $1 ")
	if err != nil {
		return fmt.Errorf("%w: %w", database.ErrCantDeleteUrl, err)
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.ErrURLNotFound
		}

		return fmt.Errorf("%w: %w", database.ErrCantDeleteUrl, err)
	}

	return nil
}
