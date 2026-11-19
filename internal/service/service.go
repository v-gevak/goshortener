package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/entity"
	"github.com/v-gevak/goshortener/internal/urlutil"
)

const maxURLLength = 2048

//go:generate go run go.uber.org/mock/mockgen -source=service.go -destination=mocks/url_mock.go -package=mocks -typed
type URLStore interface {
	SetIfAbsent(ctx context.Context, originalURL string) (*entity.URL, error)
	SetBatchIfAbsent(ctx context.Context, originalURLs []string) ([]*entity.URL, error)
	Get(ctx context.Context, key string) (*entity.URL, error)
	Ping(ctx context.Context) error
}

type Settings struct {
	BaseURL string
}

type Service struct {
	store    URLStore
	settings Settings
}

func New(store URLStore, settings Settings) (*Service, error) {
	if err := validateBaseURL(settings.BaseURL); err != nil {
		return nil, err
	}

	return &Service{store: store, settings: settings}, nil
}

func (s *Service) Shorten(ctx context.Context, originalURL string) (string, error) {
	normalizedURL, err := normalizeOriginalURL(originalURL)
	if err != nil {
		return "", err
	}

	storedURL, storeErr := s.store.SetIfAbsent(ctx, normalizedURL)
	if storeErr != nil && !errors.Is(storeErr, apperrors.ErrConflict) {
		return "", storeErr
	}

	shortURL, err := url.JoinPath(s.settings.BaseURL, storedURL.Key)
	if err != nil {
		return "", err
	}

	return shortURL, storeErr
}

func (s *Service) ShortenBatch(ctx context.Context, originalURLs []string) ([]string, error) {
	normalizedURLs := make([]string, len(originalURLs))
	for i, originalURL := range originalURLs {
		normalizedURL, err := normalizeOriginalURL(originalURL)
		if err != nil {
			return nil, fmt.Errorf("batch item [%d]: %w", i, err)
		}
		normalizedURLs[i] = normalizedURL
	}

	storedURLs, err := s.store.SetBatchIfAbsent(ctx, normalizedURLs)
	if err != nil {
		return nil, err
	}

	shortURLs := make([]string, len(storedURLs))
	for i, storedURL := range storedURLs {
		shortURLs[i], err = url.JoinPath(s.settings.BaseURL, storedURL.Key)
		if err != nil {
			return nil, err
		}
	}

	return shortURLs, nil
}

func (s *Service) Expand(ctx context.Context, key string) (string, error) {
	storedEntity, err := s.store.Get(ctx, key)

	if err != nil {
		return "", err
	}

	return storedEntity.Original, nil
}

func (s *Service) Ping(ctx context.Context) error {
	return s.store.Ping(ctx)
}

func normalizeOriginalURL(originalURL string) (string, error) {
	normalizedURL, err := urlutil.NormalizeHTTPURL(originalURL)
	if err != nil {
		return "", apperrors.NewValidationError("url", err.Error())
	}

	if len(normalizedURL) > maxURLLength {
		return "", apperrors.NewValidationError("url", fmt.Sprintf("must not exceed %d bytes after normalization", maxURLLength))
	}

	return normalizedURL, nil
}

func validateBaseURL(rawURL string) error {
	u, err := urlutil.ValidateHTTPURL(rawURL)
	if err != nil {
		return err
	}

	if port := u.Port(); port != "" {
		if err := validatePort(port); err != nil {
			return err
		}
	}

	if escapedPath := u.EscapedPath(); escapedPath != "" && path.Clean(escapedPath) != "/" {
		return errors.New("base URL path must resolve to /")
	}

	if u.RawQuery != "" || u.Fragment != "" {
		return errors.New("query parameters and fragments are not allowed")
	}

	return nil
}

func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("invalid port %q: must be between 1 and 65535", port)
	}

	return nil
}
