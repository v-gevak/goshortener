package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore_SetGet(t *testing.T) {
	t.Run("success stores and returns url", func(t *testing.T) {
		store := New()
		key := "ABCD1234"
		originalURL := "https://www.google.com"

		ok := store.SetIfAbsent(key, originalURL)

		assert.True(t, ok)

		got, ok := store.Get(key)
		assert.True(t, ok)
		assert.Equal(t, originalURL, got)
	})

	t.Run("fails if url doesn't exist", func(t *testing.T) {
		store := New()

		_, ok := store.Get("missing")
		assert.False(t, ok)
	})

}
