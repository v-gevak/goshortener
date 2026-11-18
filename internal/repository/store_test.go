package repository

import (
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/entity"
)

func TestStorage(t *testing.T) {
	const (
		findByOriginal = "SELECT short_key, original_url FROM links WHERE original_url = $1"
		findByKey      = "SELECT short_key, original_url FROM links WHERE short_key = $1"
		nextID         = "SELECT nextval(pg_get_serial_sequence('links', 'id'))"
		insertURL      = "INSERT INTO links (id, short_key, original_url) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING RETURNING short_key, original_url"
		originalURL    = "https://example.com/page"
	)
	expectInsert := func(mock sqlmock.Sqlmock, id int64, key, original string) {
		mock.ExpectQuery(findByOriginal).WithArgs(original).
			WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}))
		mock.ExpectQuery(nextID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
		mock.ExpectQuery(insertURL).WithArgs(id, key, original).
			WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow(key, original))
	}

	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		run       func(*testing.T, *Store)
	}{
		{
			name: "stores and returns url",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectInsert(mock, 1, "1", originalURL)
				mock.ExpectQuery(findByKey).WithArgs("1").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow("1", originalURL))
			},
			run: func(t *testing.T, store *Store) {
				const key = "1"

				want := entity.URL{Key: key, Original: originalURL}
				stored, err := store.SetIfAbsent(t.Context(), originalURL)
				require.NoError(t, err)
				assert.Equal(t, &want, stored)

				gotURL, err := store.Get(t.Context(), key)
				require.NoError(t, err)
				assert.Equal(t, &want, gotURL)
			},
		},
		{
			name: "fails if url doesn't exist",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(findByKey).WithArgs("missing").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}))
			},
			run: func(t *testing.T, store *Store) {
				gotURL, err := store.Get(t.Context(), "missing")
				assert.ErrorIs(t, err, apperrors.ErrNotFound)
				assert.Empty(t, gotURL)
			},
		},
		{
			name: "returns existing entry when URL is already stored",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectInsert(mock, 1, "1", originalURL)
				mock.ExpectQuery(findByOriginal).WithArgs(originalURL).
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow("1", originalURL))
				mock.ExpectQuery(findByKey).WithArgs("1").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow("1", originalURL))
				mock.ExpectQuery(findByKey).WithArgs("2").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}))
			},
			run: func(t *testing.T, store *Store) {
				const firstKey = "1"
				const secondKey = "2"

				want := entity.URL{Key: firstKey, Original: originalURL}
				stored, err := store.SetIfAbsent(t.Context(), originalURL)
				require.NoError(t, err)
				assert.Equal(t, &want, stored)

				stored, err = store.SetIfAbsent(t.Context(), originalURL)
				require.NoError(t, err)
				assert.Equal(t, &want, stored)

				gotURL, err := store.Get(t.Context(), firstKey)
				require.NoError(t, err)
				assert.Equal(t, &want, gotURL)

				_, err = store.Get(t.Context(), secondKey)
				assert.ErrorIs(t, err, apperrors.ErrNotFound)
			},
		},
		{
			name: "different URLs have different keys and remain available",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectInsert(mock, 1, "1", originalURL)
				expectInsert(mock, 2, "2", "https://example.com/other")
				mock.ExpectQuery(findByKey).WithArgs("1").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow("1", originalURL))
				mock.ExpectQuery(findByKey).WithArgs("2").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow("2", "https://example.com/other"))
			},
			run: func(t *testing.T, store *Store) {
				first, err := store.SetIfAbsent(t.Context(), originalURL)
				require.NoError(t, err)
				second, err := store.SetIfAbsent(t.Context(), "https://example.com/other")
				require.NoError(t, err)
				assert.NotEqual(t, first.Key, second.Key)

				got, err := store.Get(t.Context(), first.Key)
				require.NoError(t, err)
				assert.Equal(t, first, got)
				got, err = store.Get(t.Context(), second.Key)
				require.NoError(t, err)
				assert.Equal(t, second, got)
			},
		},
	}

	for _, kind := range []string{"in-memory", "file", "db"} {
		t.Run(kind, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					var store *Store
					if kind == "db" {
						db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
						require.NoError(t, err)
						store = &Store{storage: &dbStorage{db: db}}
						tt.setupMock(mock)
						mock.ExpectClose()
						t.Cleanup(func() {
							assert.NoError(t, mock.ExpectationsWereMet())
						})
					} else {
						settings := StoreSettings{}
						if kind == "file" {
							settings.FilePath = filepath.Join(t.TempDir(), "storage.jsonl")
						}
						var err error
						store, err = New(settings)
						require.NoError(t, err)
					}

					t.Cleanup(func() {
						assert.NoError(t, store.Close())
					})

					tt.run(t, store)
				})
			}
		})
	}
}
