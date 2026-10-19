package handler

import (
	"io"
	"log"
	"net/http"
)

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	shortURL := h.service.BuildShortURL(string(body), r.Host)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		log.Printf("write response error: %v", err)
	}
}
