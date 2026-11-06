package handler

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v-gevak/goshortener/internal/logger"
)

func TestWithLogging(t *testing.T) {
	log := &mockLogger{}

	h := withLogging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("https://www.goshortener.com/ABCD1234"))
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://www.example.com"))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "https://www.goshortener.com/ABCD1234", w.Body.String())

	require.Equal(t, 1, log.infoCalls)
	assert.Equal(t, "HTTP request completed", log.message)

	assert.Contains(t, log.fields, logger.String("method", http.MethodPost))
	assert.Contains(t, log.fields, logger.String("request URI", "/"))
	assert.Contains(t, log.fields, logger.Int("status", http.StatusCreated))
	assert.Contains(t, log.fields, logger.Int("size", 36))
	assert.True(t, slices.ContainsFunc(
		log.fields,
		func(field logger.Field) bool { return field.Key == "duration" }),
		"duration field is missing",
	)
}
