package repository

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/entity"
)

func TestProviders(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, *Store)
	}{
		{
			name: "stores and returns url",
			run: func(t *testing.T, store *Store) {
				const key = "ABCD1234"
				const originalURL = "https://example.com/page"

				want := entity.URL{Key: key, Original: originalURL}
				stored, err := store.SetIfAbsent(&want)
				require.NoError(t, err)
				assert.Equal(t, &want, stored)

				gotURL, err := store.Get(key)
				require.NoError(t, err)
				assert.Equal(t, &want, gotURL)
			},
		},
		{
			name: "returns error if key already exists",
			run: func(t *testing.T, store *Store) {
				const key = "ABCD1234"
				const originalURL = "https://a.com/page"

				want := entity.URL{Key: key, Original: originalURL}
				stored, err := store.SetIfAbsent(&want)
				require.NoError(t, err)
				assert.Equal(t, &want, stored)

				_, err = store.SetIfAbsent(&entity.URL{Key: key, Original: "https://b.com/page"})
				require.ErrorIs(t, err, apperrors.ErrConflict)

				gotURL, err := store.Get(key)
				require.NoError(t, err)
				assert.Equal(t, &want, gotURL)
			},
		},
		{
			name: "fails if url doesn't exist",
			run: func(t *testing.T, store *Store) {
				gotURL, err := store.Get("missing")
				assert.ErrorIs(t, err, apperrors.ErrNotFound)
				assert.Empty(t, gotURL)
			},
		},
		{
			name: "returns existing entry when URL is already stored",
			run: func(t *testing.T, store *Store) {
				const firstKey = "ABCD1234"
				const secondKey = "BCDE1234"
				const originalURL = "https://example.com/page"

				want := entity.URL{Key: firstKey, Original: originalURL}
				stored, err := store.SetIfAbsent(&want)
				require.NoError(t, err)
				assert.Equal(t, &want, stored)

				stored, err = store.SetIfAbsent(&entity.URL{Key: secondKey, Original: originalURL})
				require.NoError(t, err)
				assert.Equal(t, &want, stored)

				gotURL, err := store.Get(firstKey)
				require.NoError(t, err)
				assert.Equal(t, &want, gotURL)

				_, err = store.Get(secondKey)
				assert.ErrorIs(t, err, apperrors.ErrNotFound)
			},
		},
	}

	for _, kind := range []string{"in-memory", "file"} {
		t.Run(kind, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					store, err := New(StoreSettings{
						Provider: kind,
						FilePath: filepath.Join(t.TempDir(), "storage.jsonl"),
					})
					require.NoError(t, err)

					t.Cleanup(func() {
						assert.NoError(t, store.Close())
					})

					tt.run(t, store)
				})
			}
		})
	}
}
