package main

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"
)

const ContentTypePlainText = "text/plain"

var urlMap = make(map[string]string)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.Handle(`/{id}`, contentTypeMiddleware(http.HandlerFunc(getRequestHandler)))
	mux.Handle(`/`, contentTypeMiddleware(http.HandlerFunc(postRequestHandler)))

	return http.ListenAndServe(`:8080`, mux)
}

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if r.Header.Get("Content-Type") != ContentTypePlainText {
		// 	http.Error(w, "Unsupported Media Type", http.StatusBadRequest)
		// 	return
		// }

		w.Header().Set("Content-Type", "text/plain")
		next.ServeHTTP(w, r)
	})
}

func postRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil || len(body) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	key := base64.StdEncoding.EncodeToString(body)[:8]
	urlMap[key] = string(body)

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte("http://" + r.Host + "/" + key))
	if err != nil {
		panic(err)
	}
}

func getRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}

	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 2 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	val, ok := urlMap[parts[1]]
	if !ok {
		http.Error(w, "No info about requested route", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", val)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
