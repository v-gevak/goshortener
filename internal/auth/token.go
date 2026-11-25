package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenLifetime = 30 * 24 * time.Hour

type Signer struct {
	key []byte
}

func NewSigner(encodedKey string) (*Signer, error) {
	if encodedKey == "" {
		return nil, errors.New("missing signing key")
	}

	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("signing key must be valid Base64: %w", err)
	}

	if len(key) < sha256.Size {
		return nil, fmt.Errorf("signing key must contain at least %d bytes after Base64 decoding", sha256.Size)
	}

	return &Signer{key: key}, nil
}

type claims struct {
	jwt.RegisteredClaims
	UserID string
}

type IssuedToken struct {
	Value     string
	UserID    string
	ExpiresAt time.Time
}

func (s *Signer) IssueToken() (IssuedToken, error) {
	userID := newUserID()
	expiresAt := jwt.NewNumericDate(time.Now().Add(tokenLifetime))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		ExpiresAt: expiresAt,
		UserID:    userID,
	})

	tokenString, err := token.SignedString(s.key)
	if err != nil {
		return IssuedToken{}, err
	}

	return IssuedToken{
		Value:     tokenString,
		UserID:    userID,
		ExpiresAt: expiresAt.Time,
	}, nil
}

func (s *Signer) ParseUserID(tokenString string) (string, error) {
	var tokenClaims claims

	_, err := jwt.ParseWithClaims(
		tokenString,
		&tokenClaims,
		func(_ *jwt.Token) (any, error) {
			return s.key, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil {
		return "", err
	}

	return tokenClaims.UserID, nil
}

func newUserID() string {
	var id [16]byte
	rand.Read(id[:])
	return hex.EncodeToString(id[:])
}
