package router

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/router/mocks"
	"go.uber.org/mock/gomock"
)

func TestWithLogging(t *testing.T) {
	log := mocks.NewMockLogger(gomock.NewController(t))
	log.EXPECT().Info("HTTP request completed",
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Do(func(_ string, fields ...logger.Field) {
		assert.Contains(t, fields, logger.String("method", http.MethodPost))
		assert.Contains(t, fields, logger.String("request URI", "/"))
		assert.Contains(t, fields, logger.Int("status", http.StatusCreated))
		assert.Contains(t, fields, logger.Int("size", 36))
		assert.True(t, slices.ContainsFunc(
			fields,
			func(field logger.Field) bool { return field.Key == "duration" }),
			"duration field is missing",
		)
	})

	h := withLogging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("https://www.goshortener.com/ABCD1234"))
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://www.example.com"))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "https://www.goshortener.com/ABCD1234", w.Body.String())
}
