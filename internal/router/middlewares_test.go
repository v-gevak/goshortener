package router

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v-gevak/goshortener/internal/auth"
	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/router/mocks"
	"go.uber.org/mock/gomock"
)

func newTestSigner(t *testing.T) *auth.Signer {
	t.Helper()
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("test", 8)))
	signer, err := auth.NewSigner(key)
	require.NoError(t, err)
	return signer
}

func TestWithUserIdentity(t *testing.T) {
	log := mocks.NewMockLogger(gomock.NewController(t))
	h := withUserIdentity(log, newTestSigner(t))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserID(r.Context())
		require.True(t, ok)
		_, _ = w.Write([]byte(userID))
	}))

	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/", nil))
	require.Len(t, first.Body.String(), 32)

	firstResponse := first.Result()
	defer firstResponse.Body.Close()
	cookies := firstResponse.Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "user_token", cookies[0].Name)

	var claims jwt.RegisteredClaims
	_, _, err := jwt.NewParser().ParseUnverified(cookies[0].Value, &claims)
	require.NoError(t, err)
	require.NotNil(t, claims.ExpiresAt)
	assert.True(t, cookies[0].Expires.Equal(claims.ExpiresAt.Time))

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.AddCookie(cookies[0])
	second := httptest.NewRecorder()
	h.ServeHTTP(second, request)
	assert.Equal(t, first.Body.String(), second.Body.String())
	assert.Empty(t, second.Header().Values("Set-Cookie"))
}

func TestWithUserIdentityRejectsEmptyCookie(t *testing.T) {
	log := mocks.NewMockLogger(gomock.NewController(t))
	h := withUserIdentity(log, newTestSigner(t))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("request with an empty identity cookie reached the handler")
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	request.AddCookie(&http.Cookie{Name: userTokenCookieName})
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Empty(t, recorder.Header().Values("Set-Cookie"))
}

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
