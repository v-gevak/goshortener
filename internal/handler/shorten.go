package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/model"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	body, err := readRequestBody(r)
	if err != nil {
		h.writeError(w, err)
		return
	}

	if len(body) == 0 {
		h.writeError(w, model.ValidationError("empty body"))
		return
	}

	shortURL, err := h.service.Shorten(string(body))
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		h.log.Error("write response error", logger.Error(err))
	}
}

func (h *Handler) shortenJSON(w http.ResponseWriter, r *http.Request) {
	body, err := readRequestBody(r)
	if err != nil {
		h.writeError(w, err)
		return
	}

	var request shortenRequest

	if err := json.Unmarshal(body, &request); err != nil {
		h.writeError(w, model.ValidationError("invalid JSON request body"))
		return
	}

	if request.URL == "" {
		h.writeError(w, model.ValidationError("missing `URL` in request body"))
		return
	}

	shortURL, err := h.service.Shorten(request.URL)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response := shortenResponse{Result: shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("encode response error", logger.Error(err))
	}
}

func readRequestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()

	return io.ReadAll(http.MaxBytesReader(nil, r.Body, 64*1024))
}
