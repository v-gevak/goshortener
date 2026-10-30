package handler

import "github.com/v-gevak/goshortener/internal/logger"

type mockURLService struct {
	shorten func(originalURL string) (string, error)
	expand  func(key string) (string, bool)
}

func (m *mockURLService) Shorten(originalURL string) (string, error) {
	return m.shorten(originalURL)
}

func (m *mockURLService) Expand(key string) (string, bool) {
	return m.expand(key)
}

type mockLogger struct {
	infoCalls  int
	errorCalls int
	message    string
	fields     []logger.Field
}

func (l *mockLogger) Info(msg string, fields ...logger.Field) {
	l.infoCalls++
	l.message = msg
	l.fields = append([]logger.Field(nil), fields...)
}

func (l *mockLogger) Error(msg string, fields ...logger.Field) {
	l.errorCalls++
	l.message = msg
	l.fields = append([]logger.Field(nil), fields...)
}
