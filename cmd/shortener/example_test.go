package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/http/handlers"
	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
	"github.com/golangTroshin/shorturl/internal/app/service"
	"github.com/golangTroshin/shorturl/internal/app/storage"
)

func ExampleAPIShortenURL() {
	store := storage.NewMemoryStore()
	svc := service.NewURLService(store)
	handler := handlers.APIShortenURL(svc)

	requestBody := map[string]string{"url": "https://example.com"}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response Body:", w.Body.String())

	// Output:
	// Status Code: 201
	// Response Body: {"result":"http://localhost:8080/EAaArVRs"}
}

func ExampleAPIPostBatchHandler() {
	store := storage.NewMemoryStore()
	svc := service.NewURLService(store)
	handler := handlers.APIPostBatchHandler(svc)

	requestBody := []map[string]string{
		{"correlation_id": "1", "original_url": "https://example.com"},
		{"correlation_id": "2", "original_url": "https://example.org"},
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	fmt.Println("Status Code:", resp.StatusCode)

	var responseBody []map[string]string
	_ = json.NewDecoder(w.Body).Decode(&responseBody)

	for _, item := range responseBody {
		fmt.Printf("Correlation ID: %s, Short URL: %s\n", item["correlation_id"], item["short_url"])
	}

	// Output:
	// Status Code: 201
	// Correlation ID: 1, Short URL: http://localhost:8080/EAaArVRs
	// Correlation ID: 2, Short URL: http://localhost:8080/UNepBeME
}

func ExampleAPIDeleteUrlsHandler() {
	store := storage.NewMemoryStore()
	svc := service.NewURLService(store)
	handler := handlers.APIDeleteUrlsHandler(svc)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "user1")
	store.Set(ctx, "https://example.com")
	store.Set(ctx, "https://example.org")

	requestBody := []string{"EAaArVRs", "UNepBeME"}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(body))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	fmt.Println("Status Code:", resp.StatusCode)

	// Output:
	// Status Code: 202
}

func ExampleShortenURL() {
	store := storage.NewMemoryStore()
	svc := service.NewURLService(store)
	handler := handlers.ShortenURL(svc)

	requestBody := []byte("https://example.com")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response Body:", w.Body.String())

	// Output:
	// Status Code: 201
	// Response Body: http://localhost:8080/EAaArVRs
}

func ExampleGetOriginalURL() {
	store := storage.NewMemoryStore()
	svc := service.NewURLService(store)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "test-user-id")

	store.Set(ctx, "https://example.com")
	handler := handlers.GetOriginalURL(svc)

	r := chi.NewRouter()
	r.Get("/{id}", handler)

	req := httptest.NewRequest(http.MethodGet, "/EAaArVRs", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Location Header:", resp.Header.Get("Location"))

	// Output:
	// Status Code: 307
	// Location Header: https://example.com
}
