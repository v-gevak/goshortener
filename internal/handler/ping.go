package handler

import "net/http"

func (h *Handler) Ping(w http.ResponseWriter, _ *http.Request) {
	err := h.service.Ping()

	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
