package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/4aykovski/url_shortener/internal/adapters/repository"
	"github.com/4aykovski/url_shortener/internal/entity"
	"github.com/4aykovski/url_shortener/pkg/random"
)

type urlRepository interface {
	SaveURL(ctx context.Context, urlToSave string, alias string, userId int) error
	GetURL(ctx context.Context, alias string) (string, error)
	GetURLsByUserId(ctx context.Context, userId int) ([]entity.Url, error)
	DeleteURL(ctx context.Context, alias string, userId int) error
}

type UrlService struct {
	urlRepository urlRepository
}

func NewUrlService(urlRepository urlRepository) *UrlService {
	return &UrlService{
		urlRepository: urlRepository,
	}
}

const aliasLength = 6

type SaveURLInput struct {
	URL    string
	Alias  string
	UserId int
}

func (s *UrlService) SaveURL(ctx context.Context, input SaveURLInput) (string, error) {
	alias := input.Alias
	if alias == "" {
		alias = random.NewRandomString(aliasLength)
	}

	if err := s.urlRepository.SaveURL(ctx, input.URL, alias, input.UserId); err != nil {
		if errors.Is(err, repository.ErrUrlExists) {
			return "", fmt.Errorf("alias already exists: %w", ErrAliasAlreadyExists)
		}

		return "", fmt.Errorf("failed to save url: %w", err)
	}

	return alias, nil
}

type GetURLInput struct {
	Alias string
}

func (s *UrlService) GetURL(ctx context.Context, input GetURLInput) (string, error) {
	resURL, err := s.urlRepository.GetURL(ctx, input.Alias)
	if err != nil {
		if errors.Is(err, repository.ErrURLNotFound) {
			return "", fmt.Errorf("url not found: %w", ErrURLNotFound)
		}

		return "", fmt.Errorf("failed to get url: %w", err)
	}
	return resURL, nil
}

type DeleteURLInput struct {
	Alias  string
	UserId int
}

func (s *UrlService) DeleteURL(ctx context.Context, input DeleteURLInput) error {
	err := s.urlRepository.DeleteURL(ctx, input.Alias, input.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrURLNotFound) {
			return fmt.Errorf("url not found: %w", ErrURLNotFound)
		}

		return fmt.Errorf("failed to delete url: %w", err)
	}

	return nil
}

type GetAllUserUrlsInput struct {
	UserId int
}

type GetAllUserUrlsOutput struct {
	Urls map[string]string
}

func (s *UrlService) GetAllUserUrls(ctx context.Context, input GetAllUserUrlsInput) (GetAllUserUrlsOutput, error) {
	urls, err := s.urlRepository.GetURLsByUserId(ctx, input.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrURLsNotFound) {
			return GetAllUserUrlsOutput{}, fmt.Errorf("user has no urls: %w", ErrUserHasNoUrls)
		}

		return GetAllUserUrlsOutput{}, fmt.Errorf("failed to get all user urls: %w", err)
	}

	var output GetAllUserUrlsOutput
	output.Urls = make(map[string]string)
	for _, url := range urls {
		output.Urls[url.Alias] = url.Url
	}

	return output, nil
}
