package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/base62"
	"github.com/v-gevak/goshortener/internal/entity"
	"github.com/v-gevak/goshortener/migrations"
)

type dbStorage struct {
	db *sql.DB
}

func newDBStorage(dsn string) (*dbStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := migrations.ApplyAll(ctx, db); err != nil {
		db.Close()
		return nil, err
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
		return nil, err
	}

	return e, nil
}

func (s *dbStorage) SetIfAbsent(ctx context.Context, originalURL string) (*entity.URL, error) {
	const findByOriginal = `
        SELECT short_key, original_url
        FROM links
        WHERE original_url = $1
    `

	e := new(entity.URL)
	err := s.db.QueryRowContext(ctx, findByOriginal, originalURL).Scan(&e.Key, &e.Original)

	if err == nil {
		return e, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var id int64
	err = s.db.QueryRowContext(ctx, `SELECT nextval(pg_get_serial_sequence('links', 'id'))`).Scan(&id)
	if err != nil {
		return nil, err
	}

	shortKey := base62.EncodeID(id)

	err = s.db.QueryRowContext(
		ctx,
		`INSERT INTO links (id, short_key, original_url)
				VALUES ($1, $2, $3)
				ON CONFLICT (original_url) DO NOTHING
				RETURNING short_key, original_url
		`,
		id, shortKey, originalURL,
	).Scan(&e.Key, &e.Original)

	if errors.Is(err, sql.ErrNoRows) {
		err = s.db.QueryRowContext(ctx, findByOriginal, originalURL).Scan(&e.Key, &e.Original)
	}

	if err != nil {
		return nil, err
	}

	return e, nil
}

func (s *dbStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *dbStorage) Close() error {
	return s.db.Close()
}
