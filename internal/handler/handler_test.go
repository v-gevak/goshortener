package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		body            string
		expectedStatus  int
		expectedBody    string
		wantShortenCall bool
		wantExpandCall  bool
	}{
		{
			name:            "POST routes to shorten",
			method:          http.MethodPost,
			body:            "https://www.google.com",
			expectedStatus:  http.StatusCreated,
			expectedBody:    "https://shortener.ru/ABCD",
			wantShortenCall: true,
		},
		{
			name:           "GET routes to expand",
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "URL not found\n",
			wantExpandCall: true,
		},
		{
			name:           "unsupported method returns 405",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "method not allowed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortenCalled := false
			expandCalled := false

			svc := &mockURLService{
				buildShortURL: func(originalURL, host string) string {
					shortenCalled = true
					return "https://shortener.ru/ABCD"
				},
				expand: func(key string) (string, bool) {
					expandCalled = true
					return "", false
				},
			}

			h := &Handler{service: svc}

			r := httptest.NewRequest(
				tt.method,
				"/",
				strings.NewReader(tt.body),
			)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, tt.wantShortenCall, shortenCalled)
			assert.Equal(t, tt.wantExpandCall, expandCalled)
		})
	}
}
