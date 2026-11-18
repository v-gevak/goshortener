package handler

import "net/http"

func (h *Handler) Expand(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	originalURL, err := h.service.Expand(r.Context(), key)

	if err != nil {
		h.writeError(w, err)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
