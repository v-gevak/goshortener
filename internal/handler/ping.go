package handler

import "net/http"

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.service.Ping(r.Context())

	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
