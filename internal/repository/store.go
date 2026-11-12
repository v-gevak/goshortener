package repository

import (
	"github.com/v-gevak/goshortener/internal/entity"
)

type storage interface {
	Get(key string) (*entity.URL, error)
	SetIfAbsent(url *entity.URL) (*entity.URL, error)
	Ping() error
	Close() error
}

type Store struct {
	storage storage
}

type StoreSettings struct {
	FilePath string
	DBDSN    string
}

func New(settings StoreSettings) (*Store, error) {
	var repo storage
	var err error

	if settings.DBDSN != "" {
		repo, err = newDBStorage(settings.DBDSN)
	} else if settings.FilePath != "" {
		repo, err = newFileStorage(settings.FilePath)
	} else {
		repo = newInMemoryStorage()
	}

	return &Store{storage: repo}, err
}

func (s *Store) Get(key string) (*entity.URL, error) {
	return s.storage.Get(key)
}

func (s *Store) SetIfAbsent(url *entity.URL) (*entity.URL, error) {
	return s.storage.SetIfAbsent(url)
}

func (s *Store) Ping() error {
	return s.storage.Ping()
}

func (s *Store) Close() error {
	return s.storage.Close()
}
