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
	mux.Handle(`/{id}`, contentTypeMiddleware(http.HandlerFunc(getHandler)))
	mux.Handle(`/`, contentTypeMiddleware(http.HandlerFunc(postHandler)))

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

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}

	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 2 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	key := parts[1]

	if val, ok := urlMap[key]; ok {
		w.Header().Set("Location", val)
	}

	w.WriteHeader(http.StatusTemporaryRedirect)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Write([]byte("http://" + r.Host + "/" + key))
}
