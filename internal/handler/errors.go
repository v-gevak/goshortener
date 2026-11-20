package handler

import (
	"errors"
	"net/http"

	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/logger"
)

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	status, message := errorResponse(err)

	if status >= http.StatusInternalServerError {
		h.log.Error("request failed", logger.Error(err))
	}

	http.Error(w, message, status)
}

func errorResponse(err error) (int, string) {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return http.StatusRequestEntityTooLarge,
			http.StatusText(http.StatusRequestEntityTooLarge)
	}

	if validationErr, ok := errors.AsType[*apperrors.ValidationError](err); ok {
		return http.StatusBadRequest, validationErr.Error()
	}

	if errors.Is(err, apperrors.ErrNotFound) {
		return http.StatusNotFound, http.StatusText(http.StatusNotFound)
	}

	return http.StatusInternalServerError,
		http.StatusText(http.StatusInternalServerError)
}
