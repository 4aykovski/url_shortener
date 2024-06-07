package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/4aykovski/url_shortener/internal/adapters/repository"
	"github.com/4aykovski/url_shortener/internal/entity"
	"github.com/lib/pq"
)

type UrlRepositoryPostgres struct {
	postgres *Postgres
}

func NewUrlRepository(pq *Postgres) *UrlRepositoryPostgres {
	return &UrlRepositoryPostgres{postgres: pq}
}

func (repo *UrlRepositoryPostgres) SaveURL(ctx context.Context, urlToSave string, alias string, userId int) error {
	const op = "database.Postgres.UrlRepository.SaveURL"

	stmt, err := repo.postgres.db.Prepare("INSERT INTO urls(url, alias, user_id) VALUES($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, urlToSave, alias, userId)
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

func (repo *UrlRepositoryPostgres) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "database.Postgres.UrlRepository.GetURL"

	stmt, err := repo.postgres.db.Prepare("SELECT url FROM urls WHERE alias=$1")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var resultUrl string
	err = stmt.QueryRowContext(ctx, alias).Scan(&resultUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resultUrl, nil
}

func (repo *UrlRepositoryPostgres) DeleteURL(ctx context.Context, alias string, userId int) error {
	const op = "database.Postgres.UrlRepository.DeleteURL"

	stmt, err := repo.postgres.db.Prepare("DELETE FROM urls WHERE alias = $1 AND user_id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, alias, userId)
	deleted, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if deleted == 0 {
		return repository.ErrURLNotFound
	}

	return nil
}

func (repo *UrlRepositoryPostgres) GetURLsByUserId(ctx context.Context, userId int) ([]entity.Url, error) {
	const op = "database.Postgres.UrlRepository.GetURLsByUserId"

	stmt, err := repo.postgres.db.Prepare("SELECT url, alias FROM urls WHERE user_id = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var urls []entity.Url
	for rows.Next() {
		var url entity.Url
		err = rows.Scan(&url.Url, &url.Alias)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		urls = append(urls, url)
	}

	if len(urls) == 0 {
		return nil, repository.ErrURLsNotFound
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return urls, nil
}
