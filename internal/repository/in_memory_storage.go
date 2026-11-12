package repository

import (
	"sync"

	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/entity"
	"github.com/v-gevak/goshortener/internal/urlutil"
)

type inMemoryStorage struct {
	urls map[string]entity.URL
	mux  sync.Mutex
}

func newInMemoryStorage() *inMemoryStorage {
	return &inMemoryStorage{
		urls: make(map[string]entity.URL),
	}
}

func (p *inMemoryStorage) Get(key string) (*entity.URL, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	v, ok := p.urls[key]

	if !ok {
		return nil, apperrors.ErrNotFound
	}

	return &v, nil
}

func (p *inMemoryStorage) SetIfAbsent(url *entity.URL) (*entity.URL, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, urlEntity := range p.urls {
		matches, err := urlutil.CompareURLs(urlEntity.Original, url.Original)
		if err != nil {
			return nil, err
		}

		if matches {
			return &urlEntity, nil
		}
	}

	if _, exists := p.urls[url.Key]; exists {
		return nil, apperrors.ErrConflict
	}

	p.urls[url.Key] = *url

	return url, nil
}

func (p *inMemoryStorage) Ping() error {
	return nil
}

func (p *inMemoryStorage) Close() error {
	return nil
}
