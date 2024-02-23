package Postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/4aykovski/learning/golang/rest/internal/repository"
	"github.com/lib/pq"
)

type UrlRepositoryPostgres struct {
	postgres *Postgres
}

func NewUrlRepository(pq *Postgres) *UrlRepositoryPostgres {
	return &UrlRepositoryPostgres{postgres: pq}
}

func (repo *UrlRepositoryPostgres) SaveURL(urlToSave string, alias string) error {
	const op = "database.Postgres.UrlRepository.SaveURL"

	stmt, err := repo.postgres.db.Prepare("INSERT INTO url(url, alias) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(urlToSave, alias)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return repository.ErrUrlExists
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *UrlRepositoryPostgres) GetURL(alias string) (string, error) {
	const op = "database.Postgres.UrlRepository.GetURL"

	stmt, err := repo.postgres.db.Prepare("SELECT url FROM url WHERE alias=$1")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var resultUrl string
	err = stmt.QueryRow(alias).Scan(&resultUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resultUrl, nil
}

func (repo *UrlRepositoryPostgres) DeleteURL(alias string) error {
	const op = "database.Postgres.UrlRepository.DeleteURL"

	stmt, err := repo.postgres.db.Prepare("DELETE FROM url WHERE alias = $1 ")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrURLNotFound
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
