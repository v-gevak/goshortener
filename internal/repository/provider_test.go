package repository

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v-gevak/goshortener/internal/storeerr"
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

				storedKey, err := store.SetIfAbsent(key, originalURL)
				require.NoError(t, err)
				assert.Equal(t, key, storedKey)

				gotURL, err := store.Get(key)
				require.NoError(t, err)
				assert.Equal(t, originalURL, gotURL)
			},
		},
		{
			name: "returns error if key already exists",
			run: func(t *testing.T, store *Store) {
				const key = "ABCD1234"
				const originalURL = "https://a.com/page"

				storedKey, err := store.SetIfAbsent(key, originalURL)
				require.NoError(t, err)
				assert.Equal(t, key, storedKey)

				_, err = store.SetIfAbsent(key, "https://b.com/page")
				require.ErrorIs(t, err, storeerr.ErrKeyExists)

				gotURL, err := store.Get(key)
				require.NoError(t, err)
				assert.Equal(t, originalURL, gotURL)
			},
		},
		{
			name: "fails if url doesn't exist",
			run: func(t *testing.T, store *Store) {
				gotURL, err := store.Get("missing")
				assert.ErrorIs(t, err, storeerr.ErrKeyNotFound)
				assert.Empty(t, gotURL)
			},
		},
		{
			name: "returns existing entry when URL is already stored",
			run: func(t *testing.T, store *Store) {
				const firstKey = "ABCD1234"
				const secondKey = "BCDE1234"
				const originalURL = "https://example.com/page"

				storedKey, err := store.SetIfAbsent(firstKey, originalURL)
				require.NoError(t, err)
				assert.Equal(t, firstKey, storedKey)

				storedKey, err = store.SetIfAbsent(secondKey, originalURL)
				require.NoError(t, err)
				assert.Equal(t, firstKey, storedKey)

				gotURL, err := store.Get(firstKey)
				require.NoError(t, err)
				assert.Equal(t, originalURL, gotURL)

				_, err = store.Get(secondKey)
				assert.ErrorIs(t, err, storeerr.ErrKeyNotFound)
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
