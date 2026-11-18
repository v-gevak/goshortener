package repository

import (
	"context"
	"sync"

	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/base62"
	"github.com/v-gevak/goshortener/internal/entity"
)

type inMemoryStorage struct {
	lastID        int64
	urls          map[string]entity.URL
	keyByOriginal map[string]string
	mux           sync.Mutex
}

func newInMemoryStorage() *inMemoryStorage {
	return &inMemoryStorage{
		lastID:        firstShortID - 1,
		urls:          make(map[string]entity.URL),
		keyByOriginal: make(map[string]string),
	}
}

func (s *inMemoryStorage) Get(_ context.Context, key string) (*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	v, ok := s.urls[key]

	if !ok {
		return nil, apperrors.ErrNotFound
	}

	return &v, nil
}

func (s *inMemoryStorage) SetIfAbsent(_ context.Context, originalURL string) (*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.getOrCreateURL(originalURL), nil
}

func (s *inMemoryStorage) getOrCreateURL(originalURL string) *entity.URL {
	if key, ok := s.keyByOriginal[originalURL]; ok {
		stored := s.urls[key]
		return &stored
	}

	id := s.lastID + 1
	key := base62.EncodeID(id)
	e := entity.URL{
		Original: originalURL,
		Key:      key,
	}

	s.urls[key] = e
	s.keyByOriginal[originalURL] = key
	s.lastID = id

	return &e
}

func (s *inMemoryStorage) SetBatchIfAbsent(_ context.Context, originalURLs []string) ([]*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	result := make([]*entity.URL, len(originalURLs))

	for i, originalURL := range originalURLs {
		result[i] = s.getOrCreateURL(originalURL)
	}

	return result, nil
}

func (s *inMemoryStorage) Ping(context.Context) error {
	return nil
}

func (s *inMemoryStorage) Close() error {
	return nil
}
