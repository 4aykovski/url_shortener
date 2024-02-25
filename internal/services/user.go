package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	tokenManager "github.com/4aykovski/learning/golang/rest/internal/lib/token-manager"
	"github.com/4aykovski/learning/golang/rest/internal/models"
)

type userRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	DeleteUserById(ctx context.Context, id string) error
	DeleteUserByLogin(ctx context.Context, login string) error
	GetUserById(ctx context.Context, id int) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUsers(ctx context.Context) ([]models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
}

type passHasher interface {
	Hash(password string) (string, error)
	CheckPassword(password string, hashedPassword string) bool
}
type refreshSessionService interface {
	CreateRefreshSession(ctx context.Context, userId int) (*tokenManager.Tokens, error)
	GetAllUserRefreshSessions(ctx context.Context, userId int) ([]models.RefreshSession, error)
	DeleteEarliestRefreshSession(ctx context.Context, sessions []models.RefreshSession) error
	DeleteRefreshSession(ctx context.Context, refreshToken string) error
	ValidateRefreshSession(ctx context.Context, refreshToken string) (int, error)
}

type UserService struct {
	userRepo              userRepository
	refreshSessionService refreshSessionService

	hasher passHasher

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type UserSignUpInput struct {
	Login    string
	Password string
}

func NewUserService(
	userRepo userRepository,
	refreshSessionService refreshSessionService,
	hasher passHasher,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *UserService {
	return &UserService{
		userRepo:              userRepo,
		refreshSessionService: refreshSessionService,
		hasher:                hasher,
		accessTokenTTL:        accessTokenTTL,
		refreshTokenTTL:       refreshTokenTTL,
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

	err = s.userRepo.CreateUser(ctx, &u)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type UserSignInInput struct {
	Login    string
	Password string
}

var ErrWrongCred = errors.New("wrong credentials")

func (s *UserService) SignIn(ctx context.Context, input UserSignInInput) (*tokenManager.Tokens, error) {
	const op = "services.user.SignIn"

	user, err := s.userRepo.GetUserByLogin(ctx, input.Login)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ok := s.hasher.CheckPassword(input.Password, user.Password)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, ErrWrongCred)
	}

	sessions, err := s.refreshSessionService.GetAllUserRefreshSessions(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(sessions) >= 5 {
		err = s.refreshSessionService.DeleteEarliestRefreshSession(ctx, sessions)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	tokens, err := s.refreshSessionService.CreateRefreshSession(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	const op = "services.user.Logout"

	err := s.refreshSessionService.DeleteRefreshSession(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *UserService) Refresh(ctx context.Context, refreshToken string) (*tokenManager.Tokens, error) {
	const op = "services.user.Refresh"

	userId, err := s.refreshSessionService.ValidateRefreshSession(ctx, refreshToken)
	if err != nil {
		return &tokenManager.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := s.refreshSessionService.CreateRefreshSession(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}
