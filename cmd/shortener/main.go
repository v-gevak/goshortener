package main

import (
	"net/http"

	"github.com/v-gevak/goshortener/internal/handler"
)

func main() {
	h := handler.New()

	if err := http.ListenAndServe(":8080", h); err != nil {
		panic(err)
	}
}
