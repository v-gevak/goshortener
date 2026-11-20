package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func newDBStorage(ctx context.Context, dsn string) (*dbStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot open database connection: %w", err)
	}

	err = migrations.ApplyAll(ctx, db)

	if err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot initialize database: %w", err)
	}

	return &dbStorage{db: db}, nil
}

func (s *dbStorage) Get(ctx context.Context, key string) (*entity.URL, error) {
	row := s.db.QueryRowContext(ctx, `SELECT short_key, original_url FROM links WHERE short_key = $1`, key)

	e := new(entity.URL)

	err := row.Scan(&e.Key, &e.Original)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("cannot find URL by short key: %w", err)
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
		return nil, fmt.Errorf("cannot begin batch transaction: %w", err)
	}

	defer tx.Rollback()

	byURL := make(map[string]*entity.URL, len(orderedURLs))
	for _, originalURL := range orderedURLs {
		stored, err := getOrCreateURL(ctx, tx, originalURL)
		if err != nil && !errors.Is(err, apperrors.ErrConflict) {
			return nil, err
		}

		byURL[originalURL] = stored
	}

	result := make([]*entity.URL, len(originalURLs))
	for i, originalURL := range originalURLs {
		result[i] = byURL[originalURL]
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("cannot commit batch transaction: %w", err)
	}

	return result, nil
}

func (s *dbStorage) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("cannot ping database: %w", err)
	}
	return nil
}

func (s *dbStorage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("cannot close database: %w", err)
	}
	return nil
}

func getOrCreateURL(ctx context.Context, conn queryRower, originalURL string) (*entity.URL, error) {
	var nextID int64
	err := conn.QueryRowContext(ctx, `SELECT nextval(pg_get_serial_sequence('links', 'id'))`).Scan(&nextID)
	if err != nil {
		return nil, fmt.Errorf("cannot get next URL ID: %w", err)
	}

	e := new(entity.URL)
	shortKey := base62.EncodeID(nextID)

	err = conn.QueryRowContext(
		ctx,
		`INSERT INTO links (id, short_key, original_url)
				VALUES ($1, $2, $3)
				ON CONFLICT (original_url) DO NOTHING
				RETURNING short_key, original_url
		`,
		nextID, shortKey, originalURL,
	).Scan(&e.Key, &e.Original)

	if err == nil {
		return e, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("cannot insert URL: %w", err)
	}

	if err := conn.QueryRowContext(ctx, findByOriginal, originalURL).Scan(&e.Key, &e.Original); err != nil {
		return nil, fmt.Errorf("cannot find existing URL after conflict: %w", err)
	}

	return e, apperrors.ErrConflict
}
