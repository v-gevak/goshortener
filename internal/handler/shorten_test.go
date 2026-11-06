package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_shorten(t *testing.T) {
	testSvc := &mockURLService{
		shorten: func(originalURL string) (string, error) {
			return "http://example.com/ABCD", nil
		},
	}

	tests := []struct {
		name                string
		body                io.ReadCloser
		expectedStatusCode  int
		expectedBody        string
		expectedContentType string
	}{
		{
			name:                "should return error `empty body`",
			body:                io.NopCloser(strings.NewReader("")),
			expectedStatusCode:  http.StatusBadRequest,
			expectedBody:        "empty body\n",
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "should return short url",
			body:                io.NopCloser(strings.NewReader("https://www.google.com")),
			expectedStatusCode:  http.StatusCreated,
			expectedBody:        "http://example.com/ABCD",
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "should return error `413`",
			body:                io.NopCloser(strings.NewReader(strings.Repeat("a", 65*1024))),
			expectedStatusCode:  http.StatusRequestEntityTooLarge,
			expectedBody:        "Request Entity Too Large\n",
			expectedContentType: "text/plain; charset=utf-8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: testSvc,
			}

			request := httptest.NewRequest(http.MethodPost, "/", tt.body)
			w := httptest.NewRecorder()

			h.shorten(w, request)

			res := w.Result()

			assert.Equal(t, tt.expectedStatusCode, res.StatusCode)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, tt.expectedContentType, res.Header.Get("Content-Type"))
		})
	}
}
