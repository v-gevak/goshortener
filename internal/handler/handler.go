package handler

import (
	"github.com/go-chi/chi/v5"
)

type URLService interface {
	Shorten(originalURL string) (string, error)
	Expand(key string) (string, bool)
}

type Handler struct {
	service URLService
}

func New(s URLService) *chi.Mux {
	router := chi.NewRouter()
	h := &Handler{service: s}

	router.Get("/{key}", h.expand)
	router.Post("/", h.shorten)

	return router
}
