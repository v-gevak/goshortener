package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/v-gevak/goshortener/internal/model"
)

type Handler struct {
	service model.URLService
}

func New(s model.URLService) *chi.Mux {
	router := chi.NewRouter()
	h := &Handler{service: s}

	router.Get("/{key}", h.expand)
	router.Post("/", h.shorten)

	return router
}
