package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Helper function to gzip compress a string
func gzipCompress(data string) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte(data)); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestGzipMiddleware_RequestDecompression(t *testing.T) {
	compressedBody, err := gzipCompress(`{"message": "Hello, world!"}`)
	if err != nil {
		t.Fatalf("failed to compress request body: %v", err)
	}

	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		expected := `{"message": "Hello, world!"}`
		if string(body) != expected {
			t.Fatalf("expected request body '%s', got '%s'", expected, body)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(compressedBody))
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", resp.Code)
	}
}

func TestGzipMiddleware_SkipCompressionForUnsupportedContentType(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, world!"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("expected Content-Encoding to not be 'gzip'")
	}

	if resp.Body.String() != "Hello, world!" {
		t.Fatalf("expected response body to be 'Hello, world!', got '%s'", resp.Body.String())
	}
}
