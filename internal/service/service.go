package service

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"

	"github.com/v-gevak/goshortener/internal/urlutil"
)

const (
	keyLength                = 8
	alphabet                 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxKeyGenerationAttempts = 100
)

var (
	ErrInvalidURL = errors.New("invalid URL")
)

type URLStore interface {
	SetIfAbsent(key string, originalURL string) bool
	Get(key string) (string, bool)
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
	if _, err := urlutil.ValidateHTTPURL(originalURL); err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	for i := 0; i < maxKeyGenerationAttempts; i++ {
		key := s.generateKey()

		if s.store.SetIfAbsent(key, originalURL) {
			return url.JoinPath(s.settings.BaseURL, key)
		}
	}

	return "", errors.New("failed to generate unique key")
}

func (s *Service) Expand(key string) (string, bool) {
	return s.store.Get(key)
}

func (s *Service) generateKey() string {
	key := make([]byte, keyLength)

	for i := range key {
		key[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(key)
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
