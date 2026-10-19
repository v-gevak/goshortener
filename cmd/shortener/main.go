package main

import (
	"net/http"

	"github.com/v-gevak/goshortener/internal/handler"
	"github.com/v-gevak/goshortener/internal/repository"
	"github.com/v-gevak/goshortener/internal/service"
)

func main() {
	store := repository.New()
	svc := service.New(store)
	h := handler.New(svc)

	if err := http.ListenAndServe(":8080", h); err != nil {
		panic(err)
	}
}
