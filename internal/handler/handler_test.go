package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v-gevak/goshortener/internal/repository"
	"github.com/v-gevak/goshortener/internal/service"
)

func TestHandler_ShortenAndExpand(t *testing.T) {
	baseURL := "http://example.com/"
	svc, err := service.New(repository.New(), service.Settings{BaseURL: baseURL})
	require.NoError(t, err)
	h := New(svc, &mockLogger{})

	const originalURL = "https://example.org/page"
	shortenResponse := httptest.NewRecorder()
	h.ServeHTTP(shortenResponse, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL)))
	require.Equal(t, http.StatusCreated, shortenResponse.Code)
	shortURL := shortenResponse.Body.String()
	require.Regexp(t, `^http://example\.com/[a-zA-Z0-9]{8}$`, shortURL)

	expandResponse := httptest.NewRecorder()
	h.ServeHTTP(expandResponse, httptest.NewRequest(http.MethodGet, shortURL, nil))
	assert.Equal(t, http.StatusTemporaryRedirect, expandResponse.Code)
	assert.Equal(t, originalURL, expandResponse.Header().Get("Location"))

}

func TestHandler_Router(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		path            string
		body            string
		expectedStatus  int
		expectedBody    string
		wantShortenCall bool
		wantExpandCall  bool
	}{
		{
			name:            "POST routes to shorten",
			method:          http.MethodPost,
			path:            "/",
			body:            "https://www.google.com",
			expectedStatus:  http.StatusCreated,
			expectedBody:    "http://example.com/ABCD",
			wantShortenCall: true,
		},
		{
			name:           "GET routes to expand",
			method:         http.MethodGet,
			path:           "/ABCD",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "URL not found\n",
			wantExpandCall: true,
		},
		{
			name:           "unsupported method returns 405",
			method:         http.MethodPut,
			path:           "/",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortenCalled := false
			expandCalled := false

			svc := &mockURLService{
				shorten: func(originalURL string) (string, error) {
					shortenCalled = true
					return "http://example.com/ABCD", nil
				},
				expand: func(key string) (string, bool) {
					expandCalled = true
					return "", false
				},
			}

			h := New(svc, &mockLogger{})

			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, tt.wantShortenCall, shortenCalled)
			assert.Equal(t, tt.wantExpandCall, expandCalled)
		})
	}
}
