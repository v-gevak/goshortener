package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/v-gevak/goshortener/internal/config"
	"github.com/v-gevak/goshortener/internal/model"
)

type Handler struct {
	service model.URLService
	config  *config.Config
}

func New(cfg *config.Config, s model.URLService) *chi.Mux {
	router := chi.NewRouter()
	h := &Handler{
		config:  cfg,
		service: s,
	}

	router.Get("/{key}", h.expand)
	router.Post("/", h.shorten)

	return router
}
