package repository

import (
	"context"

	"github.com/v-gevak/goshortener/internal/entity"
)

const firstShortID = 62 * 62 * 62 * 62

type storage interface {
	Get(ctx context.Context, key string) (*entity.URL, error)
	GetBatchByUser(ctx context.Context, userID string) ([]*entity.URL, error)
	SetIfAbsent(ctx context.Context, originalURL, userID string) (*entity.URL, error)
	SetBatchIfAbsent(ctx context.Context, originalURLs []string, userID string) ([]*entity.URL, error)
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

func New(ctx context.Context, settings StoreSettings) (*Store, error) {
	var repo storage
	var err error

	if settings.DBDSN != "" {
		repo, err = newDBStorage(ctx, settings.DBDSN)
	} else if settings.FilePath != "" {
		repo, err = newFileStorage(settings.FilePath)
	} else {
		repo = newInMemoryStorage()
	}

	if err != nil {
		return nil, err
	}

	return &Store{storage: repo}, nil
}

func (s *Store) Get(ctx context.Context, key string) (*entity.URL, error) {
	return s.storage.Get(ctx, key)
}

func (s *Store) GetBatchByUser(ctx context.Context, userID string) ([]*entity.URL, error) {
	return s.storage.GetBatchByUser(ctx, userID)
}

func (s *Store) SetIfAbsent(ctx context.Context, originalURL, userID string) (*entity.URL, error) {
	return s.storage.SetIfAbsent(ctx, originalURL, userID)
}

func (s *Store) SetBatchIfAbsent(ctx context.Context, originalURLs []string, userID string) ([]*entity.URL, error) {
	return s.storage.SetBatchIfAbsent(ctx, originalURLs, userID)
}

func (s *Store) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

func (s *Store) Close() error {
	return s.storage.Close()
}
