package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/v-gevak/goshortener/internal/storeerr"
)

func TestHandler_expand(t *testing.T) {
	tests := []struct {
		name               string
		key                string
		expand             func(key string) (string, error)
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name: "should redirect when url exists",
			key:  "ABCD1234",
			expand: func(key string) (string, error) {
				return "https://www.google.com", nil
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "https://www.google.com",
		},
		{
			name: "should return 404 when url not found",
			key:  "zzzzzzzz",
			expand: func(key string) (string, error) {
				return "", storeerr.ErrKeyNotFound
			},
			expectedStatusCode: http.StatusNotFound,
			expectedLocation:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockURLService{
				expand: func(key string) (string, error) {
					assert.Equal(t, tt.key, key)
					return tt.expand(key)
				},
			}
			logger := &mockLogger{}
			h := New(svc, logger)

			request := httptest.NewRequest(http.MethodGet, "/"+tt.key, nil)
			request.SetPathValue("key", tt.key)

			w := httptest.NewRecorder()
			h.Expand(w, request)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedLocation, w.Header().Get("Location"))
		})
	}
}
