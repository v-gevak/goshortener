package service

import (
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
	SetIfAbsent(originalURL string) (*entity.URL, error)
	Get(key string) (*entity.URL, error)
	Ping() error
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

func (s *Service) Shorten(originalURL string) (string, error) {
	normalizedURL, err := urlutil.NormalizeHTTPURL(originalURL)
	if err != nil {
		return "", &apperrors.ValidationError{Field: "url", Message: err.Error()}
	}

	if len(normalizedURL) > maxURLLength {
		return "", &apperrors.ValidationError{
			Field:   "url",
			Message: fmt.Sprintf("must not exceed %d bytes after normalization", maxURLLength),
		}
	}

	storedURL, err := s.store.SetIfAbsent(normalizedURL)

	if err != nil {
		return "", err
	}

	return url.JoinPath(s.settings.BaseURL, storedURL.Key)
}

func (s *Service) Expand(key string) (string, error) {
	storedEntity, err := s.store.Get(key)

	if err != nil {
		return "", err
	}

	return storedEntity.Original, nil
}

func (s *Service) Ping() error {
	return s.store.Ping()
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
