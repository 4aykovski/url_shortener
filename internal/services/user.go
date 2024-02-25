package services

import (
	"context"
	"errors"
	"fmt"
	"time"

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

type tokenManager interface {
	NewJWT(userId string, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
}

type refreshSessionService interface {
	createRefreshSession(ctx context.Context, userId int, refreshToken string) error
	getAllUserRefreshSessions(ctx context.Context, userId int) ([]models.RefreshSession, error)
	deleteEarliestRefreshSession(ctx context.Context, sessions []models.RefreshSession) error
}

type UserService struct {
	userRepo              userRepository
	refreshSessionService refreshSessionService

	hasher       passHasher
	tokenManager tokenManager

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type UserSignUpInput struct {
	Login    string
	Password string
}

func NewUserService(
	userRepo userRepository,
	hasher passHasher,
	tokenManager tokenManager,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
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

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

var ErrWrongCred = errors.New("wrong credentials")

func (s *UserService) SignIn(ctx context.Context, input UserSignInInput) (*Tokens, error) {
	const op = "services.user.SignIn"

	user, err := s.userRepo.GetUserByLogin(ctx, input.Login)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ok := s.hasher.CheckPassword(input.Password, user.Password)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, ErrWrongCred)
	}

	sessions, err := s.refreshSessionService.getAllUserRefreshSessions(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(sessions) >= 5 {
		err = s.refreshSessionService.deleteEarliestRefreshSession(ctx, sessions)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	tokens, err := s.createTokensPair(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (s *UserService) createTokensPair(ctx context.Context, userId int) (*Tokens, error) {
	const op = "services.user.createTokensPair"

	accessToken, err := s.tokenManager.NewJWT(string(rune(userId)), s.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.refreshSessionService.createRefreshSession(ctx, userId, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens := Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return &tokens, nil
}
