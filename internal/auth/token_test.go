package auth

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestNewSignerRejectsInvalidKey(t *testing.T) {
	for _, tt := range []struct {
		name    string
		key     string
		wantErr string
	}{
		{name: "missing", wantErr: "missing signing key"},
		{name: "invalid Base64", key: "invalid!", wantErr: "valid Base64"},
		{
			name:    "too short",
			key:     base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 31))),
			wantErr: "at least 32 bytes",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSigner(tt.key)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestSignerVerifiesTokens(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32)))
	issuer, err := NewSigner(key)
	require.NoError(t, err)
	token, err := issuer.IssueToken()
	require.NoError(t, err)

	verifier, err := NewSigner(key)
	require.NoError(t, err)
	userID, err := verifier.ParseUserID(token.Value)
	require.NoError(t, err)
	require.Equal(t, token.UserID, userID)

	otherKey := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("b", 32)))
	otherVerifier, err := NewSigner(otherKey)
	require.NoError(t, err)
	_, err = otherVerifier.ParseUserID(token.Value)
	require.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)
}
