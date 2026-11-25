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

func (s *inMemoryStorage) GetBatchByUser(ctx context.Context, userID string) ([]*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	var entities []*entity.URL

	for _, url := range s.urls {
		if url.UserID == userID {
			entities = append(entities, &url)
		}
	}

	return entities, nil
}

func (s *inMemoryStorage) SetIfAbsent(_ context.Context, originalURL, userID string) (*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.getOrCreateURL(originalURL, userID)
}

func (s *inMemoryStorage) getOrCreateURL(originalURL, userID string) (*entity.URL, error) {
	if key, ok := s.keyByOriginal[originalURL]; ok {
		stored := s.urls[key]
		return &stored, apperrors.ErrConflict
	}

	id := s.lastID + 1
	key := base62.EncodeID(id)
	e := entity.URL{
		Original: originalURL,
		Key:      key,
		UserID:   userID,
	}

	s.urls[key] = e
	s.keyByOriginal[originalURL] = key
	s.lastID = id

	return &e, nil
}

func (s *inMemoryStorage) SetBatchIfAbsent(_ context.Context, originalURLs []string, userID string) ([]*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	result := make([]*entity.URL, len(originalURLs))

	for i, originalURL := range originalURLs {
		result[i], _ = s.getOrCreateURL(originalURL, userID)
	}

	return result, nil
}

func (s *inMemoryStorage) Ping(context.Context) error {
	return nil
}

func (s *inMemoryStorage) Close() error {
	return nil
}
