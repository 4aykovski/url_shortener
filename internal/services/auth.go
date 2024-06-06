package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/4aykovski/url_shortener/internal/adapters/repository"
	"github.com/4aykovski/url_shortener/internal/entity"
	tokenManager "github.com/4aykovski/url_shortener/pkg/manager/token"
)

type userRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	DeleteUserById(ctx context.Context, id string) error
	DeleteUserByLogin(ctx context.Context, login string) error
	GetUserById(ctx context.Context, id int) (*entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
	GetUsers(ctx context.Context) ([]entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
}

type passHasher interface {
	Hash(password string) (string, error)
	CheckPassword(password string, hashedPassword string) bool
}
type refreshSessionService interface {
	CreateRefreshSession(ctx context.Context, userId int) (*tokenManager.Tokens, error)
	GetAllUserRefreshSessions(ctx context.Context, userId int) ([]entity.RefreshSession, error)
	DeleteEarliestRefreshSession(ctx context.Context, sessions []entity.RefreshSession) error
	DeleteRefreshSession(ctx context.Context, refreshToken string) error
	ValidateRefreshSession(ctx context.Context, refreshToken string) (int, error)
}

type AuthService struct {
	userRepo              userRepository
	refreshSessionService refreshSessionService

	hasher passHasher

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(
	userRepo userRepository,
	refreshSessionService refreshSessionService,
	hasher passHasher,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:              userRepo,
		refreshSessionService: refreshSessionService,
		hasher:                hasher,
		accessTokenTTL:        accessTokenTTL,
		refreshTokenTTL:       refreshTokenTTL,
	}
}

type AuthSignUpInput struct {
	Login    string
	Password string
}

func (s *AuthService) SignUp(ctx context.Context, input AuthSignUpInput) error {
	const op = "services.user.SignUp"

	hashedPassword, err := s.hasher.Hash(input.Password)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	u := entity.User{
		Login:    input.Login,
		Password: hashedPassword,
	}

	err = s.userRepo.CreateUser(ctx, &u)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type AuthSignInInput struct {
	Login    string
	Password string
}

var ErrWrongCred = errors.New("wrong credentials")

func (s *AuthService) SignIn(ctx context.Context, input AuthSignInInput) (*tokenManager.Tokens, error) {
	const op = "services.user.SignIn"

	user, err := s.getUserWithCreds(ctx, input.Login, input.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.removeExcessRefreshSession(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := s.refreshSessionService.CreateRefreshSession(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	const op = "services.user.Logout"

	err := s.refreshSessionService.DeleteRefreshSession(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*tokenManager.Tokens, error) {
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

// getUserWithCreds checks if credentials are valid. If it's valid returns user, otherwise returns nil and error
func (s *AuthService) getUserWithCreds(ctx context.Context, login, password string) (*entity.User, error) {
	const op = "services.user.getUserWithCreds"

	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrWrongCred)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ok := s.hasher.CheckPassword(password, user.Password)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, ErrWrongCred)
	}

	return user, nil
}

func (s *AuthService) removeExcessRefreshSession(ctx context.Context, userId int) error {
	const op = "services.user.removeExcessRefreshSession"

	sessions, err := s.refreshSessionService.GetAllUserRefreshSessions(ctx, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if len(sessions) >= 5 {
		err = s.refreshSessionService.DeleteEarliestRefreshSession(ctx, sessions)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
