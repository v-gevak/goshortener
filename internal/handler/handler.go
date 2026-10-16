package handler

import (
	"net/http"
)

type Handler struct {
	urls map[string]string
}

func New() *Handler {
	return &Handler{
		urls: make(map[string]string),
	}
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
