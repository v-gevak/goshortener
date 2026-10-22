package handler

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/v-gevak/goshortener/internal/service"
)

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		log.Printf("failed to read request body: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	shortURL, err := h.service.Shorten(string(body))

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			log.Printf("failed to shorten URL: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		log.Printf("write response error: %v", err)
	}
}
