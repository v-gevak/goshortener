package handler

import (
	"context"

	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/service"
)

//go:generate go run go.uber.org/mock/mockgen -source=handler.go -destination=mocks/handler_mock.go -package=mocks -typed
type URLService interface {
	Shorten(ctx context.Context, originalURL, userID string) (string, error)
	ShortenBatch(ctx context.Context, originalURLs []string, userID string) ([]string, error)
	Expand(ctx context.Context, key string) (string, error)
	ExpandBatchByUser(ctx context.Context, userID string) ([]service.UserURL, error)
	Ping(ctx context.Context) error
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
