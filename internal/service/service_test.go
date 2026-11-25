package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/entity"
	"github.com/v-gevak/goshortener/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestService_Shorten(t *testing.T) {
	const userID = "user-1"
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
			const storedKey = "aB1"

			store := mocks.NewMockURLStore(gomock.NewController(t))
			if !tt.wantErr {
				store.EXPECT().SetIfAbsent(t.Context(), tt.originalURL, userID).
					Return(&entity.URL{Key: storedKey, Original: tt.originalURL}, nil)
			}

			svc, err := New(store, Settings{BaseURL: baseURL})
			require.NoError(t, err)

			shortURL, err := svc.Shorten(t.Context(), tt.originalURL, userID)

			if tt.wantErr {
				var validationErr *apperrors.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, "url", validationErr.Field)
				assert.NotEmpty(t, validationErr.Message)
				require.Empty(t, shortURL)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, baseURL+"/"+storedKey, shortURL)
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
			store := mocks.NewMockURLStore(gomock.NewController(t))
			_, err := New(store, Settings{BaseURL: tt.baseURL})

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
			wantErr:     apperrors.ErrNotFound,
			originalURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := mocks.NewMockURLStore(gomock.NewController(t))
			var stored *entity.URL
			if tt.wantErr == nil {
				stored = &entity.URL{Key: tt.key, Original: tt.originalURL}
			}
			store.EXPECT().Get(t.Context(), tt.key).Return(stored, tt.wantErr)

			svc, err := New(store, Settings{BaseURL: "http://example.com"})
			require.NoError(t, err)

			url, err := svc.Expand(t.Context(), tt.key)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.originalURL, url)
		})
	}
}

func TestService_ShortenBatch(t *testing.T) {
	const userID = "user-1"
	const (
		firstURL  = "https://example.com"
		secondURL = "https://example.org/page/"
	)
	tests := []struct {
		name           string
		originalURLs   []string
		normalizedURLs []string
		wantErr        bool
	}{
		{
			name:           "normalizes equivalent URLs without changing the order",
			originalURLs:   []string{"HTTPS://EXAMPLE.COM:443/", secondURL, firstURL},
			normalizedURLs: []string{firstURL, secondURL, firstURL},
		},
		{
			name:         "invalid second URL prevents storage",
			originalURLs: []string{"https://example.com", "invalid-url"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const baseURL = "http://example.com"

			store := mocks.NewMockURLStore(gomock.NewController(t))
			if !tt.wantErr {
				store.EXPECT().
					SetBatchIfAbsent(t.Context(), tt.normalizedURLs, userID).
					Return([]*entity.URL{
						{Key: "1", Original: firstURL},
						{Key: "2", Original: secondURL},
						{Key: "1", Original: firstURL},
					}, nil)
			}

			svc, err := New(store, Settings{BaseURL: baseURL})
			require.NoError(t, err)

			shortURLs, err := svc.ShortenBatch(t.Context(), tt.originalURLs, userID)

			if tt.wantErr {
				var validationErr *apperrors.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, "url", validationErr.Field)
				assert.NotEmpty(t, validationErr.Message)
				require.Empty(t, shortURLs)
				assert.Contains(t, err.Error(), "batch item [1]")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, []string{
				baseURL + "/1",
				baseURL + "/2",
				baseURL + "/1",
			}, shortURLs)
		})
	}
}
