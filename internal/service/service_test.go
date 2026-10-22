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

func (m *mockURLStore) SetIfAbsent(key, value string) bool {
	m.urls[key] = value
	return true
}

func (m *mockURLStore) Get(key string) (string, bool) {
	value, ok := m.urls[key]
	return value, ok
}

func TestService_Shorten(t *testing.T) {
	tests := []struct {
		name        string
		originalURL string
		wantErr     bool
	}{
		{
			name:        "valid url",
			originalURL: "https://www.google.com",
			wantErr:     false,
		},
		{
			name:        "invalid url",
			originalURL: "invalid-url",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const baseURL = "http://example.com"

			store := newMockURLStore()
			svc, err := New(store, Settings{BaseURL: baseURL})
			require.NoError(t, err)

			shortURL, err := svc.Shorten(tt.originalURL)

			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, shortURL)
				return
			}

			require.NoError(t, err)
			require.True(t, strings.HasPrefix(shortURL, baseURL+"/"))

			key := strings.TrimPrefix(shortURL, baseURL+"/")
			assert.Len(t, key, keyLength)

			for _, char := range key {
				assert.True(t, strings.ContainsRune(alphabet, char))
			}

			stored, ok := store.Get(key)
			require.True(t, ok)
			assert.Equal(t, tt.originalURL, stored)

		})
	}
}

func TestService_New_InvalidBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{name: "valid", baseURL: "http://example.com", wantErr: false},
		{name: "unsupported scheme", baseURL: "ftp://example.com", wantErr: true},
		{name: "empty host", baseURL: "http://", wantErr: true},
		{name: "with query", baseURL: "http://example.com?x=1", wantErr: true},
		{name: "port out of range", baseURL: "http://example.com:99999", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(newMockURLStore(), Settings{BaseURL: tt.baseURL})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
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
			svc, err := New(mockStore, Settings{BaseURL: "http://example.com"})
			require.NoError(t, err)

			if tt.exists {
				mockStore.SetIfAbsent(tt.key, tt.originalURL)
			}

			url, ok := svc.Expand(tt.key)

			assert.Equal(t, tt.exists, ok)
			assert.Equal(t, tt.originalURL, url)
		})
	}

}
