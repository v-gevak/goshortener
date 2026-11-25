package handler

import (
	"encoding/json"
	"net/http"

	"github.com/v-gevak/goshortener/internal/auth"
	"github.com/v-gevak/goshortener/internal/dto"
	"github.com/v-gevak/goshortener/internal/logger"
)

func (h *Handler) Expand(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	originalURL, err := h.service.Expand(r.Context(), key)

	if err != nil {
		h.writeError(w, err)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *Handler) ExpandBatchByUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	userID, ok := auth.UserID(r.Context())

	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	links, err := h.service.ExpandBatchByUser(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	if len(links) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]dto.ExpandBatchResponse, len(links))
	for i, link := range links {
		response[i] = dto.ExpandBatchResponse(link)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("encode response error", logger.Error(err))
	}
}
