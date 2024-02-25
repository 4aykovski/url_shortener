package Postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/4aykovski/learning/golang/rest/internal/models"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
)

type RefreshSessionRepositoryPostgres struct {
	postgres *Postgres
}

func NewRefreshSessionRepository(postgres *Postgres) *RefreshSessionRepositoryPostgres {
	return &RefreshSessionRepositoryPostgres{
		postgres: postgres,
	}
}

func (repo *RefreshSessionRepositoryPostgres) CreateRefreshSession(ctx context.Context, refreshSession *models.RefreshSession) error {
	const op = "database.Postgres.RefreshSessionRepository.CreateRefreshSession"

	stmt, err := repo.postgres.db.Prepare(`
		INSERT INTO refresh_sessions(user_id, refresh_token, expires_in) 
		VALUES($1, $2, $3)`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(
		ctx,
		refreshSession.UserId,
		refreshSession.RefreshToken,
		refreshSession.ExpiresIn,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *RefreshSessionRepositoryPostgres) DeleteRefreshSession(ctx context.Context, token string) error {
	const op = "database.Postgres.RefreshSessionRepository.DeleteRefreshSession"

	stmt, err := repo.postgres.db.Prepare("DELETE FROM refresh_sessions WHERE refresh_token = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, token)
	deleted, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if deleted == 0 {
		return repository.ErrURLNotFound
	}

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *RefreshSessionRepositoryPostgres) UpdateRefreshSession(ctx context.Context, refreshSession *models.RefreshSession) error {
	const op = "database.Postgres.RefreshSessionRepository.UpdateRefreshSession"

	stmt, err := repo.postgres.db.Prepare("UPDATE refresh_sessions SET refresh_token = $1, expires_in = $2 WHERE id = $4")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(
		ctx,
		refreshSession.RefreshToken,
		refreshSession.ExpiresIn,
		refreshSession.Id,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *RefreshSessionRepositoryPostgres) GetRefreshSession(ctx context.Context, refreshTokenId int) (*models.RefreshSession, error) {
	const op = "database.Postgres.RefreshSessionRepository.GetRefreshSession"

	stmt, err := repo.postgres.db.Prepare("SELECT id, user_id, refresh_token, expires_in FROM refresh_sessions WHERE id = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var refreshSession models.RefreshSession
	err = stmt.QueryRowContext(ctx, refreshTokenId).Scan(
		&refreshSession.Id,
		&refreshSession.UserId,
		&refreshSession.RefreshToken,
		&refreshSession.ExpiresIn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrRefreshSessionNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &refreshSession, nil
}

func (repo *RefreshSessionRepositoryPostgres) GetUserRefreshSessions(ctx context.Context, userId int) ([]models.RefreshSession, error) {
	const op = "database.Postgres.RefreshSessionRepository.GetUserRefreshSessions"

	stmt, err := repo.postgres.db.Prepare("SELECT id, user_id, refresh_token, expires_in FROM refresh_sessions WHERE user_id=$1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrRefreshSessionsNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var refreshSessions []models.RefreshSession
	for rows.Next() {
		var refreshSession models.RefreshSession
		err = rows.Scan(
			&refreshSession.Id,
			&refreshSession.UserId,
			&refreshSession.RefreshToken,
			&refreshSession.ExpiresIn,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		refreshSessions = append(refreshSessions, refreshSession)
	}

	return refreshSessions, nil
}
