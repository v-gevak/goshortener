package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SetIfAbsent(t *testing.T) {
	t.Run("stores and returns url", func(t *testing.T) {
		store := New()
		key := "ABCD1234"
		originalURL := "https://www.google.com"

		err := store.SetIfAbsent(key, originalURL)
		require.NoError(t, err)

		got, ok := store.Get(key)
		assert.True(t, ok)
		assert.Equal(t, originalURL, got)
	})

	t.Run("returns error if key already exists", func(t *testing.T) {
		store := New()
		key := "ABCD1234"

		require.NoError(t, store.SetIfAbsent(key, "https://a.com"))

		err := store.SetIfAbsent(key, "https://b.com")
		assert.Error(t, err)

		got, ok := store.Get(key)
		assert.True(t, ok)
		assert.Equal(t, "https://a.com", got)
	})

	t.Run("fails if url doesn't exist", func(t *testing.T) {
		store := New()

		_, ok := store.Get("missing")
		assert.False(t, ok)
	})
}
