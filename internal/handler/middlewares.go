package handler

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/model"
)

func withLogging(log Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

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

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func withGzipResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")

		if !acceptsGzip(r.Header.Values("Accept-Encoding")) {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzip.NewWriter(w)
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipRequestBody struct {
	*gzip.Reader
}

func (r *gzipRequestBody) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)

	if err != nil && !errors.Is(err, io.EOF) {
		return n, model.ValidationError("invalid gzip body")
	}

	return n, err
}

func withGzipRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodings := parseContentEncodings(r.Header.Values("Content-Encoding"))

		if (len(encodings) == 0) || (len(encodings) == 1 && encodings[0] == "identity") {
			next.ServeHTTP(w, r)
			return
		}

		if len(encodings) != 1 || encodings[0] != "gzip" {
			w.Header().Add("Accept-Encoding", "gzip")
			http.Error(w, "unsupported Content-Encoding", http.StatusUnsupportedMediaType)
			return
		}

		zr, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "invalid gzip body", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()
		r.Body = &gzipRequestBody{Reader: zr}
		r.ContentLength = -1
		r.Header.Del("Content-Length")
		r.Header.Del("Content-Encoding")

		next.ServeHTTP(w, r)
	})
}

func acceptsGzip(values []string) bool {
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), "gzip") {
			return true
		}
	}
	return false
}

func parseContentEncodings(values []string) []string {
	var encodings []string

	for _, value := range values {
		for encoding := range strings.SplitSeq(value, ",") {
			encoding = strings.ToLower(strings.TrimSpace(encoding))
			if encoding != "" {
				encodings = append(encodings, encoding)
			}
		}
	}

	return encodings
}

func withGzip(next http.Handler) http.Handler {
	return withGzipRequest(withGzipResponse(next))
}
