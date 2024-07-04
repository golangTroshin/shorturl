package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/handlers"
	"github.com/golangTroshin/shorturl/internal/app/stores"
)

func main() {
	if err := config.ParseFlags(); err != nil {
		log.Fatalf("error ocured while parsing flags: %v", err)
		os.Exit(1)
	}

	store := stores.NewURLStore()

	if err := http.ListenAndServe(config.Options.FlagServiceAddress, Router(store)); err != nil {
		log.Fatalf("failed to start server: %v", err)
		os.Exit(1)
	}
}

func Router(store *stores.URLStore) chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.PostRequestHandler(store))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handlers.GetRequestHandler(store))
		})
	})

	return r
}
