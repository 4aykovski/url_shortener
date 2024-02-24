package Postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/4aykovski/learning/golang/rest/internal/models"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
	"github.com/lib/pq"
)

type UserRepositoryPostgres struct {
	postgres *Postgres
}

func NewUserRepository(postgres *Postgres) *UserRepositoryPostgres {
	return &UserRepositoryPostgres{postgres: postgres}
}

func (repo *UserRepositoryPostgres) CreateUser(ctx context.Context, user *models.User) error {
	const op = "database.Postgres.UserRepository.CreateUser"

	stmt, err := repo.postgres.db.Prepare("INSERT INTO users(login, password) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, user.Login, user.Password)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return repository.ErrUserExists
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *UserRepositoryPostgres) DeleteUserById(ctx context.Context, id string) error {
	const op = "database.Postgres.UserRepository.DeleteUser"

	stmt, err := repo.postgres.db.Prepare("DELETE FROM users WHERE id = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, id)
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

func (repo *UserRepositoryPostgres) DeleteUserByLogin(ctx context.Context, login string) error {
	const op = "database.Postgres.UserRepository.DeleteUser"

	stmt, err := repo.postgres.db.Prepare("DELETE FROM users WHERE login = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, login)
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

func (repo *UserRepositoryPostgres) GetUserById(ctx context.Context, id int) (*models.User, error) {
	const op = "database.Postgres.UserRepository.GetUserById"

	stmt, err := repo.postgres.db.Prepare("SELECT id, login, password FROM users WHERE id = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var user models.User
	err = stmt.QueryRowContext(ctx, id).Scan(&user.Id, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (repo *UserRepositoryPostgres) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	const op = "database.Postgres.UserRepository.GetUserByLogin"

	stmt, err := repo.postgres.db.Prepare("SELECT id, login, password FROM users WHERE login = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var user models.User
	err = stmt.QueryRowContext(ctx, login).Scan(&user.Id, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (repo *UserRepositoryPostgres) GetUsers(ctx context.Context) ([]models.User, error) {
	const op = "database.Postgres.UserRepository.GetUsers"

	stmt, err := repo.postgres.db.Prepare("SELECT id, login, password FROM users")
	if err != nil {
		return nil, nil
	}

	var users []models.User
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUsersNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.Id, &user.Login, &user.Password)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (repo *UserRepositoryPostgres) UpdateUser(ctx context.Context, user *models.User) error {
	const op = "database.Postgres.UserRepository.UpdateUser"

	stmt, err := repo.postgres.db.Prepare("UPDATE users SET login = $1, password = $2 WHERE id = $3")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, user.Login, user.Password, user.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
