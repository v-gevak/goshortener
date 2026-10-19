package service

import (
	"math/rand"

	"github.com/v-gevak/goshortener/internal/model"
)

const (
	keyLength = 8
	alphabet  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Service struct {
	store model.URLStore
}

func New(store model.URLStore) *Service {
	return &Service{store: store}
}

func (s *Service) Shorten(originalURL string) string {
	key := s.generateUniqueKey()
	s.store.Set(key, originalURL)

	return key
}

func (s *Service) Expand(key string) (string, bool) {
	return s.store.Get(key)
}

func (s *Service) generateUniqueKey() string {
	for {
		key := make([]byte, keyLength)
		for i := range key {
			key[i] = alphabet[rand.Intn(len(alphabet))]
		}

		keyAsString := string(key)

		if !s.store.Exists(keyAsString) {
			return keyAsString
		}
	}
}
