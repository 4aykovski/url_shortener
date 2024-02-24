package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/4aykovski/learning/golang/rest/internal/models"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
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

type refreshSessionRepository interface {
	CreateRefreshSession(ctx context.Context, refreshSession *models.RefreshSession) error
	DeleteRefreshSession(ctx context.Context, id int) error
	UpdateRefreshSession(ctx context.Context, refreshSession *models.RefreshSession) error
	GetRefreshSession(ctx context.Context, refreshTokenId int) (*models.RefreshSession, error)
	GetUserRefreshSessions(ctx context.Context, userId int) ([]models.RefreshSession, error)
}

type tokenManager interface {
	NewJWT(userId string, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
}

type UserService struct {
	userRepo           userRepository
	refreshSessionRepo refreshSessionRepository

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
	refreshSessionRepo refreshSessionRepository,
	hasher passHasher,
	tokenManager tokenManager,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *UserService {
	return &UserService{
		userRepo:           userRepo,
		refreshSessionRepo: refreshSessionRepo,
		hasher:             hasher,
		tokenManager:       tokenManager,
		accessTokenTTL:     accessTokenTTL,
		refreshTokenTTL:    refreshTokenTTL,
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

	sessions, err := s.getAllUserRefreshSessions(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(sessions) >= 5 {
		err = s.deleteEarliestRefreshSession(ctx, sessions)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	tokens, err := s.createRefreshSession(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (s *UserService) createRefreshSession(ctx context.Context, userId int) (*Tokens, error) {
	const op = "services.user.createRefreshSession"

	accessToken, err := s.tokenManager.NewJWT(string(rune(userId)), s.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	session := models.RefreshSession{
		UserId:       userId,
		RefreshToken: refreshToken,
		ExpiresIn:    time.Now().Add(s.refreshTokenTTL),
	}
	err = s.refreshSessionRepo.CreateRefreshSession(ctx, &session)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens := Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return &tokens, nil
}

func (s *UserService) getAllUserRefreshSessions(ctx context.Context, userId int) ([]models.RefreshSession, error) {
	const op = "services.user.getAllUserRefreshSessions"

	sessions, err := s.refreshSessionRepo.GetUserRefreshSessions(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshSessionsNotFound) {
			return []models.RefreshSession{}, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sessions, nil
}

func (s *UserService) deleteEarliestRefreshSession(ctx context.Context, sessions []models.RefreshSession) error {
	const op = "services.user.deleteEarliestRefreshSession"

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ExpiresIn.Before(sessions[j].ExpiresIn)
	})

	err := s.refreshSessionRepo.DeleteRefreshSession(ctx, sessions[0].Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
