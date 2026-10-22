package repository

import (
	"sync"
)

type Store struct {
	urls map[string]string
	mux  sync.RWMutex
}

func New() *Store {
	return &Store{urls: make(map[string]string)}
}

func (s *Store) Get(key string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	v, ok := s.urls[key]
	return v, ok
}

func (s *Store) SetIfAbsent(key, originalURL string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.urls[key]; exists {
		return false
	}

	s.urls[key] = originalURL
	return true
}
