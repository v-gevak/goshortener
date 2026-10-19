package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) expand(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	originalURL, ok := h.service.Expand(key)

	if !ok {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
