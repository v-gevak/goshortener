package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/v-gevak/goshortener/internal/handler"
	"github.com/v-gevak/goshortener/internal/logger"
)

//go:generate go run go.uber.org/mock/mockgen -source=router.go -destination=mocks/router_mock.go -package=mocks -typed
type Logger interface {
	Info(msg string, fields ...logger.Field)
}

func New(h *handler.Handler, log Logger) http.Handler {
	router := chi.NewRouter()

	router.Use(withLogging(log))
	router.Use(withGzip)

	router.Get("/ping", h.Ping)
	router.Get("/favicon.ico", http.NotFound)
	router.Get("/{key}", h.Expand)
	router.Post("/", h.Shorten)
	router.Route("/api", func(r chi.Router) {
		r.Post("/shorten", h.ShortenJSON)
	})

	return router
}
