package main

import (
	"log"
	"net/http"

	"github.com/v-gevak/goshortener/internal/config"
	"github.com/v-gevak/goshortener/internal/handler"
	"github.com/v-gevak/goshortener/internal/repository"
	"github.com/v-gevak/goshortener/internal/service"
)

func main() {
	cfg := config.New()

	store := repository.New()
	svc, err := service.New(store, serviceSettings(cfg))
	if err != nil {
		log.Fatal(err)
	}

	h := handler.New(svc)

	log.Printf("Listening on %s\n", cfg.ServerSettings.Address)
	if err := http.ListenAndServe(cfg.ServerSettings.Address, h); err != nil {
		log.Fatal(err)
	}
}

func serviceSettings(cfg *config.Config) service.Settings {
	return service.Settings{
		BaseURL: cfg.ServiceSettings.BaseURL,
	}
}
