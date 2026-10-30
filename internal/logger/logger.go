package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type Field = zap.Field

func String(key, value string) Field {
	return zap.String(key, value)
}

func Error(err error) Field {
	return zap.Error(err)
}

func Duration(key string, duration time.Duration) Field {
	return zap.Duration(key, duration)
}

func Int(key string, val int) Field {
	return zap.Int(key, val)
}

type Logger struct {
	log *zap.Logger
}

type Settings struct {
	Level         string
	EnableFile    bool
	EnableConsole bool
	FilePath      string
}

func New(settings Settings) (*Logger, error) {
	cfg := zap.NewProductionConfig()
	if err := cfg.Level.UnmarshalText([]byte(settings.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", settings.Level, err)
	}

	cfg.OutputPaths = nil

	if settings.EnableConsole {
		cfg.OutputPaths = append(cfg.OutputPaths, "stderr")
	}

	if settings.EnableFile {
		if settings.FilePath == "" {
			return nil, fmt.Errorf("log file path is required when EnableFile=true")
		}

		dir := filepath.Dir(settings.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create log directory %q: %w", dir, err)
		}

		cfg.OutputPaths = append(cfg.OutputPaths, settings.FilePath)
	}

	cfg.ErrorOutputPaths = cfg.OutputPaths

	z, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{log: z}, nil
}

func (l *Logger) Sync() error {
	return l.log.Sync()
}

func (l *Logger) Debug(msg string, fields ...Field) {
	l.log.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.log.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.log.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.log.Error(msg, fields...)
}
