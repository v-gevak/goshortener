package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/v-gevak/goshortener/internal/config"
	"github.com/v-gevak/goshortener/internal/handler"
	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/repository"
	"github.com/v-gevak/goshortener/internal/service"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot load config:", err)
		return err
	}

	log, err := logger.New(logSettings(cfg))
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot create logger:", err)
		return err
	}

	defer func() {
		if err := log.Sync(); err != nil {
			fmt.Fprintln(os.Stderr, "sync logger:", err)
		}
	}()

	store := repository.New()
	svc, err := service.New(store, serviceSettings(cfg))
	if err != nil {
		log.Error("cannot create service", logger.Error(err))
		return err
	}

	h := handler.New(svc, log)

	log.Info("starting HTTP server", logger.String("address", cfg.ServerSettings.Address))
	if err := http.ListenAndServe(cfg.ServerSettings.Address, h); err != nil {
		log.Error("HTTP server failed", logger.Error(err))
		return err
	}

	return nil
}

func serviceSettings(cfg *config.Config) service.Settings {
	return service.Settings{
		BaseURL: cfg.ServiceSettings.BaseURL,
	}
}

func logSettings(cfg *config.Config) logger.Settings {
	return logger.Settings{
		Level:         cfg.LogSettings.Level,
		EnableFile:    cfg.LogSettings.EnableFile,
		EnableConsole: cfg.LogSettings.EnableConsole,
		FilePath:      cfg.LogSettings.FilePath,
	}
}
