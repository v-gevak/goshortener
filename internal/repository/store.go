package repository

import (
	"sync"

	"github.com/v-gevak/goshortener/internal/storeerr"
)

type Store struct {
	urls map[string]string
	mux  sync.Mutex
}

func New() *Store {
	return &Store{urls: make(map[string]string)}
}

func (s *Store) Get(key string) (string, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	v, ok := s.urls[key]
	return v, ok
}

func (s *Store) SetIfAbsent(key, originalURL string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.urls[key]; exists {
		return storeerr.ErrKeyExists
	}

	s.urls[key] = originalURL
	return nil
}
