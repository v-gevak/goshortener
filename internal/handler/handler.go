package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/v-gevak/goshortener/internal/logger"
)

type URLService interface {
	Shorten(originalURL string) (string, error)
	Expand(key string) (string, bool)
}

type Logger interface {
	Error(msg string, fields ...logger.Field)
	Info(msg string, fields ...logger.Field)
}

type Handler struct {
	service URLService
	log     Logger
}

func New(s URLService, log Logger) *chi.Mux {
	router := chi.NewRouter()
	h := &Handler{service: s, log: log}

	router.Use(withLogging(log))

	router.Get("/{key}", h.expand)
	router.Post("/", h.shorten)

	return router
}
