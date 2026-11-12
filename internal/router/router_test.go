package router

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/handler"
	handlermocks "github.com/v-gevak/goshortener/internal/handler/mocks"
	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/repository"
	"github.com/v-gevak/goshortener/internal/router/mocks"
	"github.com/v-gevak/goshortener/internal/service"
	"go.uber.org/mock/gomock"
)

func TestRouter_ShortenAndExpand(t *testing.T) {
	baseURL := "http://example.com/"
	store, err := repository.New(repository.StoreSettings{})
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	svc, err := service.New(store, service.Settings{BaseURL: baseURL})
	require.NoError(t, err)
	ctrl := gomock.NewController(t)
	log := mocks.NewMockLogger(ctrl)
	log.EXPECT().Info("HTTP request completed",
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Times(2)
	handlerLog := handlermocks.NewMockLogger(ctrl)
	h := New(handler.New(svc, handlerLog), log)

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

func TestRouter_Routes(t *testing.T) {
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
			expectedBody:   "Not Found\n",
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
			ctrl := gomock.NewController(t)
			svc := handlermocks.NewMockURLService(ctrl)
			if tt.wantShortenCall {
				svc.EXPECT().Shorten(tt.body).Return("http://example.com/ABCD", nil)
			}
			if tt.wantExpandCall {
				svc.EXPECT().Expand(strings.TrimPrefix(tt.path, "/")).Return("", apperrors.ErrNotFound)
			}

			log := mocks.NewMockLogger(ctrl)
			log.EXPECT().Info("HTTP request completed",
				logger.String("request URI", tt.path),
				logger.String("method", tt.method),
				logger.Int("status", tt.expectedStatus),
				gomock.Any(), gomock.Any(),
			)
			handlerLog := handlermocks.NewMockLogger(ctrl)
			h := New(handler.New(svc, handlerLog), log)

			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestRouter_Compression(t *testing.T) {
	const originalURL = "http://google.com"
	const shortURL = "http://example.com/ABCD"
	const jsonResponse = `{"result":"http://example.com/ABCD"}`

	compress := func(body []byte) []byte {
		t.Helper()

		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		_, err := writer.Write(body)

		require.NoError(t, err)
		require.NoError(t, writer.Close())
		return buf.Bytes()
	}

	jsonBody := []byte(`{"url":"http://google.com"}`)
	compressedJSON := compress(jsonBody)

	tests := []struct {
		name             string
		path             string
		body             []byte
		contentEncoding  string
		acceptEncoding   string
		expectedStatus   int
		expectedBody     string
		expectedEncoding string
	}{
		{
			name:            "gzip request and uncompressed response",
			path:            "/api/shorten",
			body:            compressedJSON,
			contentEncoding: "gzip",
			expectedStatus:  http.StatusCreated,
			expectedBody:    jsonResponse,
		},
		{
			name:             "uncompressed request and gzip response",
			path:             "/api/shorten",
			body:             jsonBody,
			acceptEncoding:   "gzip",
			expectedStatus:   http.StatusCreated,
			expectedBody:     jsonResponse,
			expectedEncoding: "gzip",
		},
		{
			name:             "gzip request and response",
			path:             "/api/shorten",
			body:             compressedJSON,
			contentEncoding:  "gzip",
			acceptEncoding:   "gzip",
			expectedStatus:   http.StatusCreated,
			expectedBody:     jsonResponse,
			expectedEncoding: "gzip",
		},
		{
			name:            "gzip request and uncompressed plain text response",
			path:            "/",
			body:            compress([]byte(originalURL)),
			contentEncoding: "gzip",
			acceptEncoding:  "gzip",
			expectedStatus:  http.StatusCreated,
			expectedBody:    shortURL,
		},
		{
			name:             "uppercase accepted encoding",
			path:             "/api/shorten",
			body:             jsonBody,
			acceptEncoding:   "GZIP",
			expectedStatus:   http.StatusCreated,
			expectedBody:     jsonResponse,
			expectedEncoding: "gzip",
		},
		{
			name:             "gzip among accepted encodings",
			path:             "/api/shorten",
			body:             jsonBody,
			acceptEncoding:   "deflate, gzip, br",
			expectedStatus:   http.StatusCreated,
			expectedBody:     jsonResponse,
			expectedEncoding: "gzip",
		},
		{
			name:            "invalid gzip header returns 400",
			path:            "/api/shorten",
			body:            []byte("not a gzip body"),
			contentEncoding: "gzip",
			expectedStatus:  http.StatusBadRequest,
			expectedBody:    "validation failed: invalid gzip body\n",
		},
		{
			name:            "truncated gzip returns uncompressed 400",
			path:            "/api/shorten",
			body:            compressedJSON[:len(compressedJSON)-4],
			contentEncoding: "gzip",
			acceptEncoding:  "gzip",
			expectedStatus:  http.StatusBadRequest,
			expectedBody:    "validation failed: invalid gzip body\n",
		},
		{
			name:            "decompressed body over limit returns uncompressed 413",
			path:            "/api/shorten",
			body:            compress(bytes.Repeat([]byte("a"), 64*1024+1)),
			contentEncoding: "gzip",
			acceptEncoding:  "gzip",
			expectedStatus:  http.StatusRequestEntityTooLarge,
			expectedBody:    "Request Entity Too Large\n",
		},
		{
			name:            "unsupported content-encoding returns 415",
			path:            "/api/shorten",
			body:            compressedJSON,
			contentEncoding: "br",
			acceptEncoding:  "gzip",
			expectedStatus:  http.StatusUnsupportedMediaType,
			expectedBody:    "unsupported Content-Encoding\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := handlermocks.NewMockURLService(ctrl)
			if tt.expectedStatus == http.StatusCreated {
				svc.EXPECT().Shorten(originalURL).Return(shortURL, nil)
			}
			log := mocks.NewMockLogger(ctrl)
			log.EXPECT().Info("HTTP request completed",
				logger.String("request URI", tt.path),
				logger.String("method", http.MethodPost),
				logger.Int("status", tt.expectedStatus),
				gomock.Any(), gomock.Any(),
			)
			handlerLog := handlermocks.NewMockLogger(ctrl)
			h := New(handler.New(svc, handlerLog), log)

			r := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewReader(tt.body))
			if tt.contentEncoding != "" {
				r.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			if tt.acceptEncoding != "" {
				r.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			require.Equal(t, tt.expectedEncoding, w.Header().Get("Content-Encoding"))

			var reader io.Reader = w.Body
			if tt.expectedEncoding == "gzip" {
				gz, err := gzip.NewReader(w.Body)
				require.NoError(t, err)
				defer gz.Close()
				reader = gz
			}
			body, err := io.ReadAll(reader)
			require.NoError(t, err)

			if tt.path == "/api/shorten" && tt.expectedStatus == http.StatusCreated {
				assert.JSONEq(t, tt.expectedBody, string(body))
			} else {
				assert.Equal(t, tt.expectedBody, string(body))
			}
		})
	}
}
