package main

import (
	"fmt"
	"net/http"

	"github.com/v-gevak/goshortener/internal/config"
	"github.com/v-gevak/goshortener/internal/handler"
	"github.com/v-gevak/goshortener/internal/repository"
	"github.com/v-gevak/goshortener/internal/service"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	store := repository.New()
	svc := service.New(store)
	h := handler.New(cfg, svc)

	fmt.Printf("Listening on %s\n", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, h); err != nil {
		panic(err)
	}
}
