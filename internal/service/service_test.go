package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v-gevak/goshortener/internal/model"
	"github.com/v-gevak/goshortener/internal/storeerr"
)

type mockURLStore struct {
	urls map[string]string
}

func newMockURLStore() *mockURLStore {
	return &mockURLStore{
		urls: make(map[string]string),
	}
}

func (m *mockURLStore) SetIfAbsent(key, value string) (string, error) {
	m.urls[key] = value
	return key, nil
}

func (m *mockURLStore) Get(key string) (string, error) {
	value, ok := m.urls[key]
	if !ok {
		return "", storeerr.ErrKeyNotFound
	}

	return value, nil
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
		{
			name:        "port without hostname",
			originalURL: "http://:8080/path",
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
				require.ErrorAs(t, err, new(model.ValidationError))
				require.Empty(t, shortURL)
				require.Empty(t, store.urls)
				return
			}

			require.NoError(t, err)
			require.True(t, strings.HasPrefix(shortURL, baseURL+"/"))

			key := strings.TrimPrefix(shortURL, baseURL+"/")
			assert.Len(t, key, keyLength)

			for _, char := range key {
				assert.True(t, strings.ContainsRune(alphabet, char))
			}

			stored, err := store.Get(key)
			require.NoError(t, err)
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
		{name: "root path", baseURL: "http://example.com/", wantErr: false},
		{name: "path prefix", baseURL: "http://example.com/short", wantErr: true},
		{name: "unsupported scheme", baseURL: "ftp://example.com", wantErr: true},
		{name: "empty host", baseURL: "http://", wantErr: true},
		{name: "port without hostname", baseURL: "http://:8080", wantErr: true},
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
		wantErr     error
		originalURL string
	}{
		{
			name:        "returns original URL when key exists",
			key:         "ABCD1234",
			wantErr:     nil,
			originalURL: "https://www.google.com",
		},
		{
			name:        "returns error when key does not exist",
			key:         "NOTFOUND",
			wantErr:     storeerr.ErrKeyNotFound,
			originalURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := newMockURLStore()
			svc, err := New(mockStore, Settings{BaseURL: "http://example.com"})
			require.NoError(t, err)

			if tt.wantErr == nil {
				_, err := mockStore.SetIfAbsent(tt.key, tt.originalURL)
				require.NoError(t, err)
			}

			url, err := svc.Expand(tt.key)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.originalURL, url)
		})
	}

}
