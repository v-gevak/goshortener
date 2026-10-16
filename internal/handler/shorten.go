package handler

import (
	"io"
	"log"
	"math/rand"
	"net/http"
)

const (
	keyLength = 8
	alphabet  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	key := h.generateKey()
	h.urls[key] = originalURL

	shortURL := "http://" + r.Host + "/" + key
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		log.Printf("write response error: %v", err)
	}
}

func (h *Handler) generateKey() string {
	for {
		key := make([]byte, keyLength)
		for i := range key {
			key[i] = alphabet[rand.Intn(len(alphabet))]
		}
		if _, exists := h.urls[string(key)]; !exists {
			return string(key)
		}
	}
}
