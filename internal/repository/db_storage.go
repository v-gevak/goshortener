package repository

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/base62"
	"github.com/v-gevak/goshortener/internal/entity"
	"github.com/v-gevak/goshortener/migrations"
)

const (
	findByOriginal = `SELECT short_key, original_url FROM links WHERE original_url = $1`
)

type dbStorage struct {
	db *sql.DB
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func newDBStorage(dsn string) (*dbStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &dbStorage{db: db}, nil
}

func (s *dbStorage) bootstrap(ctx context.Context) error {
	return migrations.ApplyAll(ctx, s.db)
}

func (s *dbStorage) Get(ctx context.Context, key string) (*entity.URL, error) {
	row := s.db.QueryRowContext(ctx, `SELECT short_key, original_url FROM links WHERE short_key = $1`, key)

	e := new(entity.URL)

	err := row.Scan(&e.Key, &e.Original)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return e, nil
}

func (s *dbStorage) SetIfAbsent(ctx context.Context, originalURL string) (*entity.URL, error) {
	return getOrCreateURL(ctx, s.db, originalURL)
}

func (s *dbStorage) SetBatchIfAbsent(ctx context.Context, originalURLs []string) ([]*entity.URL, error) {
	orderedURLs := slices.Clone(originalURLs)
	slices.Sort(orderedURLs)
	orderedURLs = slices.Compact(orderedURLs)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	byURL := make(map[string]*entity.URL, len(orderedURLs))
	for _, originalURL := range orderedURLs {
		stored, err := getOrCreateURL(ctx, tx, originalURL)
		if err != nil {
			return nil, err
		}

		byURL[originalURL] = stored
	}

	result := make([]*entity.URL, len(originalURLs))
	for i, originalURL := range originalURLs {
		result[i] = byURL[originalURL]
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *dbStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *dbStorage) Close() error {
	return s.db.Close()
}

func getOrCreateURL(ctx context.Context, conn queryRower, originalURL string) (*entity.URL, error) {
	e := new(entity.URL)
	err := conn.QueryRowContext(ctx, findByOriginal, originalURL).Scan(&e.Key, &e.Original)

	if err == nil {
		return e, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var id int64
	err = conn.QueryRowContext(ctx, `SELECT nextval(pg_get_serial_sequence('links', 'id'))`).Scan(&id)
	if err != nil {
		return nil, err
	}

	shortKey := base62.EncodeID(id)

	err = conn.QueryRowContext(
		ctx,
		`INSERT INTO links (id, short_key, original_url)
				VALUES ($1, $2, $3)
				ON CONFLICT (original_url) DO NOTHING
				RETURNING short_key, original_url
		`,
		id, shortKey, originalURL,
	).Scan(&e.Key, &e.Original)

	if errors.Is(err, sql.ErrNoRows) {
		err = conn.QueryRowContext(ctx, findByOriginal, originalURL).Scan(&e.Key, &e.Original)
	}

	if err != nil {
		return nil, err
	}

	return e, nil
}
