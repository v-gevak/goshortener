package repository

import (
	"context"

	"github.com/v-gevak/goshortener/internal/entity"
)

const firstShortID = 62 * 62 * 62 * 62

type storage interface {
	Get(ctx context.Context, key string) (*entity.URL, error)
	SetIfAbsent(ctx context.Context, originalURL string) (*entity.URL, error)
	Ping(ctx context.Context) error
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

func (s *Store) Bootstrap(ctx context.Context) error {
	if db, ok := s.storage.(*dbStorage); ok {
		return db.bootstrap(ctx)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (*entity.URL, error) {
	return s.storage.Get(ctx, key)
}

func (s *Store) SetIfAbsent(ctx context.Context, originalURL string) (*entity.URL, error) {
	return s.storage.SetIfAbsent(ctx, originalURL)
}

func (s *Store) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

func (s *Store) Close() error {
	return s.storage.Close()
}
