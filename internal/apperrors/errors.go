package apperrors

import (
	"errors"
	"strings"
)

var (
	ErrConflict = errors.New("short key already exists")
	ErrNotFound = errors.New("url not found")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	parts := []string{"validation failed"}
	if e.Field != "" {
		parts = append(parts, e.Field)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	return strings.Join(parts, ": ")
}
