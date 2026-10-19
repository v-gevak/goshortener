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

		store.Set(key, originalURL)

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

func TestStore_Exists(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "returns true for existing key",
			key:      "ABCD1234",
			expected: true,
		},
		{
			name:     "returns false for missing key",
			key:      "zzzzzzzz",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := New()
			store.Set("ABCD1234", "https://www.google.com")

			assert.Equal(t, tt.expected, store.Exists(tt.key))
		})
	}
}
