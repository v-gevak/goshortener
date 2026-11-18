package repository

import (
	"context"
	"sync"

	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/base62"
	"github.com/v-gevak/goshortener/internal/entity"
)

type inMemoryStorage struct {
	lastID int64
	urls   map[string]entity.URL
	mux    sync.Mutex
}

func newInMemoryStorage() *inMemoryStorage {
	return &inMemoryStorage{
		urls: make(map[string]entity.URL),
	}
}

func (p *inMemoryStorage) Get(_ context.Context, key string) (*entity.URL, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	v, ok := p.urls[key]

	if !ok {
		return nil, apperrors.ErrNotFound
	}

	return &v, nil
}

func (p *inMemoryStorage) SetIfAbsent(_ context.Context, originalURL string) (*entity.URL, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, urlEntity := range p.urls {
		if urlEntity.Original == originalURL {
			return &urlEntity, nil
		}
	}

	id := p.lastID + 1
	key := base62.EncodeID(id)
	e := entity.URL{
		Original: originalURL,
		Key:      key,
	}

	p.urls[key] = e
	p.lastID = id

	return &e, nil
}

func (p *inMemoryStorage) Ping(context.Context) error {
	return nil
}

func (p *inMemoryStorage) Close() error {
	return nil
}
