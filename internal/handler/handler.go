package handler

import (
	"github.com/v-gevak/goshortener/internal/logger"
)

//go:generate go run go.uber.org/mock/mockgen -source=handler.go -destination=mocks/handler_mock.go -package=mocks -typed
type URLService interface {
	Shorten(originalURL string) (string, error)
	Expand(key string) (string, error)
}

type Logger interface {
	Error(msg string, fields ...logger.Field)
	Info(msg string, fields ...logger.Field)
}

type Handler struct {
	service URLService
	log     Logger
}

func New(s URLService, log Logger) *Handler {
	return &Handler{
		service: s,
		log:     log,
	}
}
