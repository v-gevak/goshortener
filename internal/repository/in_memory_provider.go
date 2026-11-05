package repository

import (
	"sync"

	"github.com/v-gevak/goshortener/internal/storeerr"
	"github.com/v-gevak/goshortener/internal/urlutil"
)

type inMemoryProvider struct {
	urls map[string]string
	mux  sync.Mutex
}

func newInMemoryProvider() *inMemoryProvider {
	return &inMemoryProvider{
		urls: map[string]string{},
	}
}

func (p *inMemoryProvider) Get(key string) (string, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	v, ok := p.urls[key]

	if !ok {
		return "", storeerr.ErrKeyNotFound
	}

	return v, nil
}

func (p *inMemoryProvider) SetIfAbsent(key, originalURL string) (string, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	for existingKey, existingURL := range p.urls {
		matches, err := urlutil.CompareURLs(existingURL, originalURL)
		if err != nil {
			return "", err
		}

		if matches {
			return existingKey, nil
		}
	}

	if _, exists := p.urls[key]; exists {
		return "", storeerr.ErrKeyExists
	}

	p.urls[key] = originalURL

	return key, nil
}

func (p *inMemoryProvider) Close() error {
	return nil
}
