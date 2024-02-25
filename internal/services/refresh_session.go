package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	tokenManager "github.com/4aykovski/learning/golang/rest/internal/lib/token-manager"
	"github.com/4aykovski/learning/golang/rest/internal/models"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
)

var (
	ErrTokenExpired = errors.New("token expired")
)

type refreshSessionRepository interface {
	CreateRefreshSession(ctx context.Context, refreshSession *models.RefreshSession) error
	DeleteRefreshSession(ctx context.Context, token string) error
	UpdateRefreshSession(ctx context.Context, refreshSession *models.RefreshSession) error
	GetRefreshSession(ctx context.Context, refreshToken string) (*models.RefreshSession, error)
	GetUserRefreshSessions(ctx context.Context, userId int) ([]models.RefreshSession, error)
}

type RefreshSessionService struct {
	refreshSessionRepo refreshSessionRepository

	tokenManager tokenManager.TokenManager

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewRefreshSessionService(
	refreshSessionRepo refreshSessionRepository,
	tokenManager tokenManager.TokenManager,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *RefreshSessionService {
	return &RefreshSessionService{
		refreshSessionRepo: refreshSessionRepo,
		tokenManager:       tokenManager,
		accessTokenTTL:     accessTokenTTL,
		refreshTokenTTL:    refreshTokenTTL,
	}
}

func (s *RefreshSessionService) CreateRefreshSession(ctx context.Context, userId int) (*tokenManager.Tokens, error) {
	const op = "services.refresh_session.CreateRefreshSession"

	tokens, err := s.tokenManager.CreateTokensPair(string(rune(userId)), s.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	session := models.RefreshSession{
		UserId:       userId,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    time.Now().Add(s.refreshTokenTTL),
	}
	err = s.refreshSessionRepo.CreateRefreshSession(ctx, &session)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (s *RefreshSessionService) GetAllUserRefreshSessions(ctx context.Context, userId int) ([]models.RefreshSession, error) {
	const op = "services.refresh_session.GetAllUserRefreshSessions"

	sessions, err := s.refreshSessionRepo.GetUserRefreshSessions(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshSessionsNotFound) {
			return []models.RefreshSession{}, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sessions, nil
}

func (s *RefreshSessionService) DeleteEarliestRefreshSession(ctx context.Context, sessions []models.RefreshSession) error {
	const op = "services.refresh_session.DeleteEarliestRefreshSession"

	// sort sessions in ascending order
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ExpiresIn.Before(sessions[j].ExpiresIn)
	})

	err := s.refreshSessionRepo.DeleteRefreshSession(ctx, sessions[0].RefreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *RefreshSessionService) DeleteRefreshSession(ctx context.Context, refreshToken string) error {
	const op = "services.refresh_session.DeleteRefreshSession"

	err := s.refreshSessionRepo.DeleteRefreshSession(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *RefreshSessionService) ValidateRefreshSession(ctx context.Context, refreshToken string) (int, error) {
	const op = "services.refresh_session.ValidateRefreshSession"

	session, err := s.refreshSessionRepo.GetRefreshSession(ctx, refreshToken)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	err = s.refreshSessionRepo.DeleteRefreshSession(ctx, session.RefreshToken)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	ok := s.isSessionExpired(session)
	if !ok {
		return -1, ErrTokenExpired
	}

	return session.UserId, nil
}

func (s *RefreshSessionService) isSessionExpired(session *models.RefreshSession) bool {
	return session.ExpiresIn.After(time.Now())
}
