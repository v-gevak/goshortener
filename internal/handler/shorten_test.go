package handler

import (
	"errors"
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
			expectedBody:        "validation failed: empty body\n",
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
				log:     &mockLogger{},
			}

			request := httptest.NewRequest(http.MethodPost, "/", tt.body)
			w := httptest.NewRecorder()

			h.Shorten(w, request)

			res := w.Result()

			assert.Equal(t, tt.expectedStatusCode, res.StatusCode)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, tt.expectedContentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestHandler_shortenJSON(t *testing.T) {
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
		svc                 *mockURLService
	}{
		{
			name:                "return error missing url in request body",
			body:                io.NopCloser(strings.NewReader(`{"a":"b"}`)),
			expectedStatusCode:  http.StatusBadRequest,
			expectedBody:        "validation failed: url: must not be empty\n",
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "return error for malformed JSON",
			body:                io.NopCloser(strings.NewReader(`{"url":`)),
			expectedStatusCode:  http.StatusBadRequest,
			expectedBody:        "validation failed: invalid JSON request body\n",
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "return `201`",
			body:                io.NopCloser(strings.NewReader(`{"url":"https://www.google.com"}`)),
			expectedStatusCode:  http.StatusCreated,
			expectedBody:        "{\"result\":\"http://example.com/ABCD\"}\n",
			expectedContentType: "application/json",
		},
		{
			name:                "return error `413`",
			body:                io.NopCloser(strings.NewReader(strings.Repeat("a", 65*1024))),
			expectedStatusCode:  http.StatusRequestEntityTooLarge,
			expectedBody:        "Request Entity Too Large\n",
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "return error `500`",
			body:                io.NopCloser(strings.NewReader(`{"url":"https://www.google.com"}`)),
			expectedStatusCode:  http.StatusInternalServerError,
			expectedBody:        "Internal Server Error\n",
			expectedContentType: "text/plain; charset=utf-8",
			svc: &mockURLService{shorten: func(originalURL string) (string, error) {
				return "", errors.New("failed to generate unique key")
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: testSvc,
				log:     &mockLogger{},
			}

			if tt.svc != nil {
				h.service = tt.svc
			}

			request := httptest.NewRequest(http.MethodPost, "/", tt.body)
			w := httptest.NewRecorder()

			h.ShortenJSON(w, request)

			res := w.Result()

			assert.Equal(t, tt.expectedStatusCode, res.StatusCode)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, tt.expectedContentType, res.Header.Get("Content-Type"))
		})
	}
}
