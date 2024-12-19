package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/middleware"
	"github.com/golangTroshin/shorturl/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestPostRequestHandler(t *testing.T) {
	store := storage.NewMemoryStore()
	handler := PostRequestHandler(store)

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("Successful URL Shortening", func(t *testing.T) {
		reqBody := []byte("https://example.com")
		resp, err := http.Post(server.URL, "text/plain", bytes.NewReader(reqBody))

		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "/")
	})

	t.Run("Empty Body", func(t *testing.T) {
		resp, err := http.Post(server.URL, "text/plain", nil)

		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestGetRequestHandler(t *testing.T) {
	store := storage.NewMemoryStore()
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "test-user")

	// Store a URL
	url, err := store.Set(ctx, "https://example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, url.ShortURL)

	// Simulate routing with chi
	router := chi.NewRouter()
	router.Get("/{id}", GetRequestHandler(store))

	// Create a request to the short URL
	req := httptest.NewRequest(http.MethodGet, "/"+url.ShortURL, nil)
	recorder := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(recorder, req)

	// Assert the response
	assert.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
	assert.Equal(t, "https://example.com", recorder.Header().Get("Location"))
}

func TestGetURLsByUserHandler(t *testing.T) {
	store := storage.NewMemoryStore()
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "test-user")
	store.Set(ctx, "https://example1.com")
	store.Set(ctx, "https://example2.com")

	handler := GetURLsByUserHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []map[string]string
	body, _ := io.ReadAll(recorder.Body)
	err := json.Unmarshal(body, &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
}

func TestDatabasePing(t *testing.T) {
	handler := DatabasePing()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
