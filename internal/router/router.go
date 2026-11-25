package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/v-gevak/goshortener/internal/auth"
	"github.com/v-gevak/goshortener/internal/handler"
	"github.com/v-gevak/goshortener/internal/logger"
)

//go:generate go run go.uber.org/mock/mockgen -source=router.go -destination=mocks/router_mock.go -package=mocks -typed
type Logger interface {
	Info(msg string, fields ...logger.Field)
	Error(msg string, fields ...logger.Field)
}

func New(h *handler.Handler, log Logger, signer *auth.Signer) http.Handler {
	router := chi.NewRouter()
	userIdentity := withUserIdentity(log, signer)

	router.Use(withLogging(log))
	router.Use(withGzip)

	router.Get("/ping", h.Ping)
	router.Get("/favicon.ico", http.NotFound)
	router.Get("/{key}", h.Expand)
	router.With(userIdentity).Post("/", h.Shorten)
	router.Route("/api", func(r chi.Router) {
		r.Use(userIdentity)

		r.Route("/user", func(r chi.Router) {
			r.Get("/urls", h.ExpandBatchByUser)
		})
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", h.ShortenJSON)
			r.Post("/batch", h.ShortenBatch)
		})
	})

	return router
}
