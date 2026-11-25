package router

import (
	"errors"
	"net/http"

	"github.com/v-gevak/goshortener/internal/auth"
	"github.com/v-gevak/goshortener/internal/logger"
)

const userTokenCookieName = "user_token"

func withUserIdentity(log Logger, signer *auth.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string
			cookie, err := r.Cookie(userTokenCookieName)

			switch {
			case errors.Is(err, http.ErrNoCookie):
				userID, err = issueUserCookie(w, signer)
			case err != nil:
				log.Error("cannot get user token", logger.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			case cookie.Value != "":
				userID, err = signer.ParseUserID(cookie.Value)
				if err != nil {
					userID, err = issueUserCookie(w, signer)
				}
			}

			if err != nil {
				log.Error("cannot issue auth token", logger.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if userID == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := auth.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func issueUserCookie(w http.ResponseWriter, signer *auth.Signer) (string, error) {
	token, err := signer.IssueToken()
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     userTokenCookieName,
		Value:    token.Value,
		Path:     "/",
		Expires:  token.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
	})
	return token.UserID, nil
}
