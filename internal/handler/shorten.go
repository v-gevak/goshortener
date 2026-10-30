package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/service"
)

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64*1024))

	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}

		h.log.Error("failed to read request body", logger.Error(err))
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
			h.log.Error("failed to shorten URL", logger.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		h.log.Error("write response error", logger.Error(err))
	}
}
