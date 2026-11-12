package repository

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/v-gevak/goshortener/internal/entity"
)

type dbStorage struct {
	db *sql.DB
}

func newDBStorage(dsn string) (*dbStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &dbStorage{db: db}, nil
}

func (s *dbStorage) Get(key string) (*entity.URL, error) {
	return nil, nil
}

func (s *dbStorage) SetIfAbsent(url *entity.URL) (*entity.URL, error) {
	return nil, nil
}

func (s *dbStorage) Ping() error {
	return s.db.PingContext(context.Background())
}

func (s *dbStorage) Close() error {
	return s.db.Close()
}
