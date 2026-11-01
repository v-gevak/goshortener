package handler

import (
	"errors"
	"net/http"

	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/model"
)

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := http.StatusText(status)
	_, tooLarge := errors.AsType[*http.MaxBytesError](err)
	_, validation := errors.AsType[model.ValidationError](err)

	switch {
	case tooLarge:
		status = http.StatusRequestEntityTooLarge
		message = http.StatusText(status)
	case validation:
		status = http.StatusBadRequest
		message = err.Error()
	default:
		h.log.Error("request failed", logger.Error(err))
	}

	http.Error(w, message, status)
}
