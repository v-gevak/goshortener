package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockURLStore struct {
	urls map[string]string
}

func newMockURLStore() *mockURLStore {
	return &mockURLStore{
		urls: make(map[string]string),
	}
}

func (m *mockURLStore) Set(key, value string) {
	m.urls[key] = value
}

func (m *mockURLStore) Get(key string) (string, bool) {
	value, ok := m.urls[key]
	return value, ok
}

func (m *mockURLStore) Exists(key string) bool {
	_, ok := m.urls[key]
	return ok
}

func TestService_BuildShortURL(t *testing.T) {
	store := newMockURLStore()
	svc := New(store)

	originalURL := "https://www.google.com"
	host := "shortener.ru"

	shortURL := svc.BuildShortURL(originalURL, host)

	prefix := "http://" + host + "/"

	require.True(t, strings.HasPrefix(shortURL, prefix))

	key := strings.TrimPrefix(shortURL, prefix)

	assert.Len(t, key, keyLength)

	for _, char := range key {
		assert.True(t, strings.ContainsRune(alphabet, char))
	}
}

func TestService_Expand(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		exists      bool
		originalURL string
	}{
		{
			name:        "returns original URL when key exists",
			key:         "ABCD1234",
			exists:      true,
			originalURL: "https://www.google.com",
		},
		{
			name:        "returns false when key does not exist",
			key:         "NOTFOUND",
			exists:      false,
			originalURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := newMockURLStore()
			svc := New(mockStore)

			if tt.exists {
				mockStore.Set(tt.key, tt.originalURL)
			}

			url, ok := svc.Expand(tt.key)

			assert.Equal(t, tt.exists, ok)
			assert.Equal(t, tt.originalURL, url)
		})
	}

}
