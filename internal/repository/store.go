package repository

import (
	"errors"

	"github.com/v-gevak/goshortener/internal/entity"
)

type storeProvider interface {
	Get(key string) (*entity.URL, error)
	SetIfAbsent(url *entity.URL) (*entity.URL, error)
	Close() error
}

type Store struct {
	provider storeProvider
}

type StoreSettings struct {
	FilePath string
	Provider string
}

func New(settings StoreSettings) (*Store, error) {
	var provider storeProvider
	var err error

	switch settings.Provider {
	case "in-memory":
		provider = newInMemoryProvider()
	case "file":
		provider, err = newFileProvider(settings.FilePath)
	default:
		return nil, errors.New("unknown store provider")
	}

	return &Store{provider: provider}, err
}

func (s *Store) Get(key string) (*entity.URL, error) {
	return s.provider.Get(key)
}

func (s *Store) SetIfAbsent(url *entity.URL) (*entity.URL, error) {
	return s.provider.SetIfAbsent(url)
}

func (s *Store) Close() error {
	return s.provider.Close()
}
