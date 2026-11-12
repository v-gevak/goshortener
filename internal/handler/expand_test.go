package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/handler/mocks"
	"go.uber.org/mock/gomock"
)

func TestHandler_expand(t *testing.T) {
	tests := []struct {
		name               string
		key                string
		originalURL        string
		serviceErr         error
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name:               "should redirect when url exists",
			key:                "ABCD1234",
			originalURL:        "https://www.google.com",
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "https://www.google.com",
		},
		{
			name:               "should return 404 when url not found",
			key:                "zzzzzzzz",
			serviceErr:         apperrors.ErrNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedLocation:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockURLService(ctrl)
			svc.EXPECT().Expand(tt.key).Return(tt.originalURL, tt.serviceErr)
			log := mocks.NewMockLogger(ctrl)
			h := New(svc, log)

			request := httptest.NewRequest(http.MethodGet, "/"+tt.key, nil)
			request.SetPathValue("key", tt.key)

			w := httptest.NewRecorder()
			h.Expand(w, request)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedLocation, w.Header().Get("Location"))
		})
	}
}
