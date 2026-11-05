package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) expand(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	originalURL, err := h.service.Expand(key)

	if err != nil {
		h.writeError(w, err)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
