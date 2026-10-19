package handler

import (
	"net/http"

	"github.com/v-gevak/goshortener/internal/model"
)

type Handler struct {
	service model.URLService
}

func New(s model.URLService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		h.shorten(w, r)
	case http.MethodGet:

		h.expand(w, r)
	default:

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
