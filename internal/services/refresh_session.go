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

type refreshSessionRepository interface {
	CreateRefreshSession(ctx context.Context, refreshSession *models.RefreshSession) error
	DeleteRefreshSession(ctx context.Context, id int) error
	UpdateRefreshSession(ctx context.Context, refreshSession *models.RefreshSession) error
	GetRefreshSession(ctx context.Context, refreshTokenId int) (*models.RefreshSession, error)
	GetUserRefreshSessions(ctx context.Context, userId int) ([]models.RefreshSession, error)
}

type RefreshSessionService struct {
	refreshSessionRepo refreshSessionRepository

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewRefreshSessionService(
	refreshSessionRepo refreshSessionRepository,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *RefreshSessionService {
	return &RefreshSessionService{
		refreshSessionRepo: refreshSessionRepo,
		accessTokenTTL:     accessTokenTTL,
		refreshTokenTTL:    refreshTokenTTL,
	}
}

func (s *RefreshSessionService) createRefreshSession(ctx context.Context, userId int, refreshToken string) error {
	const op = "services.refresh_session.CreateRefreshSession"

	session := models.RefreshSession{
		UserId:       userId,
		RefreshToken: refreshToken,
		ExpiresIn:    time.Now().Add(s.refreshTokenTTL),
	}
	err := s.refreshSessionRepo.CreateRefreshSession(ctx, &session)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *RefreshSessionService) getAllUserRefreshSessions(ctx context.Context, userId int) ([]models.RefreshSession, error) {
	const op = "services.refresh_session.getAllUserRefreshSessions"

	sessions, err := s.refreshSessionRepo.GetUserRefreshSessions(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshSessionsNotFound) {
			return []models.RefreshSession{}, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sessions, nil
}

func (s *RefreshSessionService) deleteEarliestRefreshSession(ctx context.Context, sessions []models.RefreshSession) error {
	const op = "services.refresh_session.deleteEarliestRefreshSession"

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ExpiresIn.Before(sessions[j].ExpiresIn)
	})

	err := s.refreshSessionRepo.DeleteRefreshSession(ctx, sessions[0].Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
