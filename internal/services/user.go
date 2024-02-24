package services

import (
	"context"
	"fmt"

	"github.com/4aykovski/learning/golang/rest/internal/models"
)

type userRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	DeleteUserById(ctx context.Context, id string) error
	DeleteUserByLogin(ctx context.Context, login string) error
	GetUserById(ctx context.Context, id int) (*models.User, error)
	GetUsers(ctx context.Context) ([]models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
}

type passHasher interface {
	Hash(password string) (string, error)
	CheckPassword(password string, hashedPassword string) bool
}

type UserService struct {
	repo   userRepository
	hasher passHasher
}

type UserSignUpInput struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=72,containsany=!*&^?#@)(-+=$_"`
}

func NewUserService(
	repo userRepository,
	hasher passHasher,
) *UserService {
	return &UserService{
		repo:   repo,
		hasher: hasher,
	}
}

func (s *UserService) SignUp(ctx context.Context, input UserSignUpInput) error {
	const op = "services.user.SignUp"

	hashedPassword, err := s.hasher.Hash(input.Password)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	u := models.User{
		Login:    input.Login,
		Password: hashedPassword,
	}

	err = s.repo.CreateUser(ctx, &u)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
