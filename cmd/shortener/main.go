package main

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/go-chi/chi"
)

const ContentTypePlainText = "text/plain"

var urlMap = make(map[string]string)

func main() {
	if err := http.ListenAndServe(`:8080`, Router()); err != nil {
		panic(err)
	}
}

func Router() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/", postRequestHandler)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", getRequestHandler)
		})
	})

	return r
}

func postRequestHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil || len(body) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	key := base64.URLEncoding.EncodeToString(body)[:8]
	urlMap[key] = string(body)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte("http://" + r.Host + "/" + key))
	if err != nil {
		panic(err)
	}
}

func getRequestHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	val, ok := urlMap[id]
	if !ok {
		http.Error(w, "No info about requested route", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", val)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
