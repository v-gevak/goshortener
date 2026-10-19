package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_expand(t *testing.T) {
	tests := []struct {
		name               string
		key                string
		expand             func(key string) (string, bool)
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name: "should redirect when url exists",
			key:  "ABCD1234",
			expand: func(key string) (string, bool) {
				return "https://www.google.com", true
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "https://www.google.com",
		},
		{
			name: "should return 404 when url not found",
			key:  "zzzzzzzz",
			expand: func(key string) (string, bool) {
				return "", false
			},
			expectedStatusCode: http.StatusNotFound,
			expectedLocation:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: &mockURLService{
					expand: tt.expand,
				},
			}

			request := httptest.NewRequest(http.MethodGet, "/"+tt.key, nil)
			w := httptest.NewRecorder()

			h.expand(w, request)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedLocation, w.Header().Get("Location"))

		})
	}
}
