package logger_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golangTroshin/shorturl/internal/app/logger"
	"github.com/stretchr/testify/assert"
)

func TestLoggingWrapper(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	var logBuffer bytes.Buffer
	originalStderr := os.Stderr
	defer func() { os.Stderr = originalStderr }()
	r, w, _ := os.Pipe()
	os.Stderr = w

	middleware := logger.LoggingWrapper(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	w.Close()
	_, _ = io.Copy(&logBuffer, r)

	resp := rec.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "OK", string(body))

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "uri /test-uri")
	assert.Contains(t, logOutput, "method GET")
	assert.Contains(t, logOutput, "status 200")
	assert.Contains(t, logOutput, "size 2")
	assert.Contains(t, logOutput, "duration")
}
