package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/auth"
	"github.com/v-gevak/goshortener/internal/dto"
	"github.com/v-gevak/goshortener/internal/logger"
)

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	body, err := readRequestBody(w, r)
	if err != nil {
		h.writeError(w, err)
		return
	}

	if len(body) == 0 {
		h.writeError(w, apperrors.NewValidationError("", "empty body"))
		return
	}

	userID, _ := auth.UserID(r.Context())

	shortURL, err := h.service.Shorten(r.Context(), string(body), userID)
	status := http.StatusCreated
	switch {
	case errors.Is(err, apperrors.ErrConflict):
		status = http.StatusConflict
	case err != nil:
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		h.log.Error("write response error", logger.Error(err))
	}
}

func (h *Handler) ShortenJSON(w http.ResponseWriter, r *http.Request) {
	body, err := readRequestBody(w, r)
	if err != nil {
		h.writeError(w, err)
		return
	}

	var request dto.ShortenRequest

	if err := json.Unmarshal(body, &request); err != nil {
		h.writeError(w, apperrors.NewValidationError("", "invalid JSON request body"))
		return
	}

	if err := request.Validate(); err != nil {
		h.writeError(w, err)
		return
	}

	userID, _ := auth.UserID(r.Context())

	shortURL, err := h.service.Shorten(r.Context(), request.URL, userID)
	status := http.StatusCreated
	switch {
	case errors.Is(err, apperrors.ErrConflict):
		status = http.StatusConflict
	case err != nil:
		h.writeError(w, err)
		return
	}

	response := dto.ShortenResponse{Result: shortURL}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("encode response error", logger.Error(err))
	}
}

func (h *Handler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1048576))
	if err != nil {
		h.writeError(w, err)
		return
	}

	var request []dto.ShortenRequestBatch

	if err := json.Unmarshal(body, &request); err != nil {
		h.writeError(w, apperrors.NewValidationError("", "invalid JSON request body"))
		return
	}

	if err := dto.ValidateShortenRequestBatch(request); err != nil {
		h.writeError(w, err)
		return
	}

	originalURLs := make([]string, len(request))
	for i, item := range request {
		originalURLs[i] = item.OriginalURL
	}

	userID, _ := auth.UserID(r.Context())

	shortURLs, err := h.service.ShortenBatch(r.Context(), originalURLs, userID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	var response = make([]dto.ShortenResponseBatch, len(request))

	for i, shortURL := range shortURLs {
		response[i] = dto.ShortenResponseBatch{
			CorrelationID: request[i].CorrelationID,
			ShortURL:      shortURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("encode response error", logger.Error(err))
	}
}

func readRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	defer r.Body.Close()

	return io.ReadAll(http.MaxBytesReader(w, r.Body, 16*1024))
}
