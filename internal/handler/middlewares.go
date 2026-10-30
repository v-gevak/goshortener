package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/v-gevak/goshortener/internal/logger"
)

func withLogging(log Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			h.ServeHTTP(ww, r)

			duration := time.Since(start)

			log.Info(
				"HTTP request completed",
				logger.String("request URI", r.RequestURI),
				logger.String("method", r.Method),
				logger.Int("status", ww.Status()),
				logger.Duration("duration", duration),
				logger.Int("size", ww.BytesWritten()),
			)
		})
	}
}
