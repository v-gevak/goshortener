package repository

import (
	"bytes"
	"encoding/json"
	"os"
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
		{
			name: "caller changes do not alter stored data",
			run: func(t *testing.T, store *Store) {
				want := entity.URL{Key: "ABCD1234", Original: "https://example.com/page"}
				input := want
				stored, err := store.SetIfAbsent(&input)
				require.NoError(t, err)
				input.Key = "changed-input"
				input.Original = "https://changed-input.example"
				stored.Key = "changed-result"
				stored.Original = "https://changed-result.example"

				got, err := store.Get(want.Key)
				require.NoError(t, err)
				require.Equal(t, &want, got)
				got.Key = "changed-get"
				got.Original = "https://changed-get.example"

				duplicate, err := store.SetIfAbsent(&entity.URL{Key: "different", Original: want.Original})
				require.NoError(t, err)
				require.Equal(t, &want, duplicate)
				duplicate.Original = "https://changed-duplicate.example"

				got, err = store.Get(want.Key)
				require.NoError(t, err)
				assert.Equal(t, &want, got)
				_, err = store.Get("different")
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

func TestFileProvider_PreservesStorageFormat(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "storage.jsonl")
	const previousRecord = `{"uuid":"550e8400-e29b-41d4-a716-446655440000","short_url":"ABCD1234","original_url":"https://example.com/page"}` + "\n"
	require.NoError(t, os.WriteFile(filePath, []byte(previousRecord), 0600))

	store, err := New(StoreSettings{Provider: "file", FilePath: filePath})
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, store.Close()) })

	want := entity.URL{Key: "ABCD1234", Original: "https://example.com/page"}
	got, err := store.Get(want.Key)
	require.NoError(t, err)
	assert.Equal(t, &want, got)

	got, err = store.SetIfAbsent(&entity.URL{Key: "unused", Original: want.Original})
	require.NoError(t, err)
	assert.Equal(t, &want, got)

	added := entity.URL{Key: "BCDE1234", Original: "https://other.example/page"}
	got, err = store.SetIfAbsent(&added)
	require.NoError(t, err)
	assert.Equal(t, &added, got)

	contents, err := os.ReadFile(filePath)
	require.NoError(t, err)
	lines := bytes.Split(bytes.TrimSpace(contents), []byte("\n"))
	require.Len(t, lines, 2)
	assert.Equal(t, previousRecord, string(lines[0])+"\n")
	var record map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(lines[1], &record))
	require.Len(t, record, 3)
	assert.JSONEq(t, `"BCDE1234"`, string(record["short_url"]))
	assert.JSONEq(t, `"https://other.example/page"`, string(record["original_url"]))
	assert.NotEmpty(t, record["uuid"])

	reopened, err := New(StoreSettings{Provider: "file", FilePath: filePath})
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, reopened.Close()) })
	for _, want := range []entity.URL{want, added} {
		got, err := reopened.Get(want.Key)
		require.NoError(t, err)
		assert.Equal(t, &want, got)
	}
}
