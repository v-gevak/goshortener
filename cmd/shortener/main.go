package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/v-gevak/goshortener/internal/config"
	"github.com/v-gevak/goshortener/internal/handler"
	"github.com/v-gevak/goshortener/internal/logger"
	"github.com/v-gevak/goshortener/internal/repository"
	"github.com/v-gevak/goshortener/internal/router"
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

	store, err := repository.New(storeSettings(cfg))
	if err != nil {
		log.Error("cannot create store", logger.Error(err))
		return err
	}

	defer func() {
		if err := store.Close(); err != nil {
			log.Error("store close error", logger.Error(err))
		}
	}()

	svc, err := service.New(store, serviceSettings(cfg))
	if err != nil {
		log.Error("cannot create service", logger.Error(err))
		return err
	}

	h := handler.New(svc, log)
	r := router.New(h, log)

	server := &http.Server{
		Handler: r,
		Addr:    cfg.ServerSettings.Address,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	serverErr := make(chan error, 1)
	go func() {
		log.Info("starting HTTP server", logger.String("address", cfg.ServerSettings.Address))
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		log.Error("HTTP server failed", logger.Error(err))
		return err
	case <-stop:
	}

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Failed to shutdown server", logger.Error(err))
		return err
	}

	log.Info("Server gracefully stopped")

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

func storeSettings(cfg *config.Config) repository.StoreSettings {
	return repository.StoreSettings{
		Provider: cfg.StoreSettings.Provider,
		FilePath: cfg.StoreSettings.FilePath,
	}
}
