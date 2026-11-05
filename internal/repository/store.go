package repository

import (
	"errors"
)

type storeProvider interface {
	Get(string) (string, error)
	SetIfAbsent(key, originalURL string) (string, error)
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

func (s *Store) Get(key string) (string, error) {
	return s.provider.Get(key)
}

func (s *Store) SetIfAbsent(key, originalURL string) (string, error) {
	return s.provider.SetIfAbsent(key, originalURL)
}

func (s *Store) Close() error {
	return s.provider.Close()
}
