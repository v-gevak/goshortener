package dto

import (
	"fmt"
	"strings"

	"github.com/v-gevak/goshortener/internal/apperrors"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

func (r ShortenRequest) Validate() error {
	if r.URL == "" {
		return &apperrors.ValidationError{Field: "url", Message: "must not be empty"}
	}

	return nil
}

type ShortenRequestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

func (r ShortenRequestBatch) Validate() error {
	if strings.TrimSpace(r.CorrelationID) == "" {
		return &apperrors.ValidationError{
			Field:   "correlation_id",
			Message: "must not be empty",
		}
	}

	return nil
}

func ValidateShortenRequestBatch(items []ShortenRequestBatch) error {
	if len(items) == 0 {
		return &apperrors.ValidationError{
			Message: "batch must not be empty",
		}
	}

	for i, item := range items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("batch item [%d]: %w", i, err)
		}
	}

	return nil
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenResponseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
