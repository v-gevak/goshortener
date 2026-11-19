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

const findByKey = "SELECT short_key, original_url FROM links WHERE short_key = $1"
const nextID = "SELECT nextval(pg_get_serial_sequence('links', 'id'))"
const insertURL = "INSERT INTO links (id, short_key, original_url) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING RETURNING short_key, original_url"

func expectInsert(mock sqlmock.Sqlmock, id int64, key, original string) {
	mock.ExpectQuery(nextID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectQuery(insertURL).WithArgs(id, key, original).
		WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow(key, original))
}

func expectConflict(mock sqlmock.Sqlmock, id int64, attemptedKey, existingKey, original string) {
	mock.ExpectQuery(nextID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectQuery(insertURL).WithArgs(id, attemptedKey, original).
		WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}))
	mock.ExpectQuery(findByOriginal).WithArgs(original).
		WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow(existingKey, original))
}

func TestStorage(t *testing.T) {
	const originalURL = "https://example.com/page"

	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		run       func(*testing.T, *Store)
	}{
		{
			name: "stores and returns url",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectInsert(mock, firstShortID, "10000", originalURL)
				mock.ExpectQuery(findByKey).WithArgs("10000").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow("10000", originalURL))
			},
			run: func(t *testing.T, store *Store) {
				const key = "10000"

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
				expectInsert(mock, firstShortID, "10000", originalURL)
				expectConflict(mock, firstShortID+1, "10001", "10000", originalURL)
				mock.ExpectQuery(findByKey).WithArgs("10000").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow("10000", originalURL))
				mock.ExpectQuery(findByKey).WithArgs("10001").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}))
			},
			run: func(t *testing.T, store *Store) {
				const firstKey = "10000"
				const secondKey = "10001"

				want := entity.URL{Key: firstKey, Original: originalURL}
				stored, err := store.SetIfAbsent(t.Context(), originalURL)
				require.NoError(t, err)
				assert.Equal(t, &want, stored)

				stored, err = store.SetIfAbsent(t.Context(), originalURL)
				require.ErrorIs(t, err, apperrors.ErrConflict)
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
				expectInsert(mock, firstShortID, "10000", originalURL)
				expectInsert(mock, firstShortID+1, "10001", "https://example.com/other")
				mock.ExpectQuery(findByKey).WithArgs("10000").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow("10000", originalURL))
				mock.ExpectQuery(findByKey).WithArgs("10001").
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow("10001", "https://example.com/other"))
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

func TestStorage_SetBatched(t *testing.T) {
	const (
		firstURL  = "https://google.com"
		secondURL = "https://yandex.ru"
		firstKey  = "10000"
		secondKey = "10001"
		thirdKey  = "10002"
	)
	expectGet := func(mock sqlmock.Sqlmock, key, original string) {
		mock.ExpectQuery(findByKey).WithArgs(key).
			WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}).AddRow(key, original))
	}

	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		run       func(*testing.T, *Store)
	}{
		{
			name: "stores batch and returns urls",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				expectInsert(mock, firstShortID, firstKey, firstURL)
				expectInsert(mock, firstShortID+1, secondKey, secondURL)
				mock.ExpectCommit()
				expectGet(mock, firstKey, firstURL)
				expectGet(mock, secondKey, secondURL)
			},
			run: func(t *testing.T, store *Store) {
				want := []*entity.URL{
					{Key: firstKey, Original: firstURL},
					{Key: secondKey, Original: secondURL},
				}
				stored, err := store.SetBatchIfAbsent(t.Context(), []string{firstURL, secondURL})
				require.NoError(t, err)
				assert.Equal(t, want, stored)

				for _, expected := range want {
					gotURL, err := store.Get(t.Context(), expected.Key)
					require.NoError(t, err)
					assert.Equal(t, expected, gotURL)
				}
			},
		},
		{
			name: "returns existing entry when URL is already stored",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				expectInsert(mock, firstShortID, firstKey, firstURL)
				mock.ExpectCommit()

				mock.ExpectBegin()
				expectConflict(mock, firstShortID+1, secondKey, firstKey, firstURL)
				expectInsert(mock, firstShortID+2, thirdKey, secondURL)
				mock.ExpectCommit()

				expectGet(mock, firstKey, firstURL)
				expectGet(mock, thirdKey, secondURL)
				mock.ExpectQuery(findByKey).WithArgs(secondKey).
					WillReturnRows(sqlmock.NewRows([]string{"short_key", "original_url"}))
			},
			run: func(t *testing.T, store *Store) {
				firstEntries := []*entity.URL{
					{Key: firstKey, Original: firstURL},
				}

				stored, err := store.SetBatchIfAbsent(t.Context(), []string{firstEntries[0].Original})
				require.NoError(t, err)
				assert.Equal(t, firstEntries, stored)

				storedKey, unusedKey := secondKey, thirdKey
				if _, isDB := store.storage.(*dbStorage); isDB {
					storedKey, unusedKey = thirdKey, secondKey
				}
				secondEntries := []*entity.URL{
					{Key: storedKey, Original: secondURL},
					{Key: firstKey, Original: firstURL},
					{Key: storedKey, Original: secondURL},
				}

				stored, err = store.SetBatchIfAbsent(t.Context(), []string{
					secondEntries[0].Original,
					secondEntries[1].Original,
					secondEntries[2].Original,
				})
				require.NoError(t, err)
				assert.Equal(t, secondEntries, stored)

				gotURL, err := store.Get(t.Context(), firstKey)
				require.NoError(t, err)
				assert.Equal(t, firstEntries[0], gotURL)

				gotURL, err = store.Get(t.Context(), storedKey)
				require.NoError(t, err)
				assert.Equal(t, secondEntries[0], gotURL)

				_, err = store.Get(t.Context(), unusedKey)
				assert.ErrorIs(t, err, apperrors.ErrNotFound)
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
