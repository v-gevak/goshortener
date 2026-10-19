package handler

import (
	"net/http"
	"strings"
)

func (h *Handler) expand(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, ok := h.service.Expand(key)

	if !ok {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
