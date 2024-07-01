package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostRequestHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		content     string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "positive test #1",
			body: "https://practicum.yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
				content:     "/aHR0cHM6",
			},
		},
		{
			name: "negative test #1",
			body: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				content:     "Empty body\n",
			},
		},
	}

	for _, tt := range tests {
		router := Router()

		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
		r.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)
		result := w.Result()

		assert.Equal(t, tt.want.code, result.StatusCode)
		assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

		resultURL, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		err = result.Body.Close()
		require.NoError(t, err)

		if tt.want.code == http.StatusCreated {
			expectedURL := "http://" + r.Host + tt.want.content
			assert.Equal(t, expectedURL, string(resultURL))
		} else {
			assert.Equal(t, tt.want.content, string(resultURL))
		}
	}
}

func TestGetRequestHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		location    string
	}
	tests := []struct {
		name       string
		requestURI string
		want       want
	}{
		{
			name:       "positive test #1",
			requestURI: "/aHR0cHM6",
			want: want{
				code:        http.StatusTemporaryRedirect,
				contentType: "text/plain",
				location:    "https://practicum.yandex.ru/",
			},
		},
		{
			name:       "negative test #1",
			requestURI: "/aHR0cHM63444",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
		{
			name:       "negative test #2",
			requestURI: "/aHR0cHM6/123/qwewqe",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
	}

	for _, tt := range tests {
		router := Router()

		r := httptest.NewRequest(http.MethodGet, tt.requestURI, nil)
		r.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)
		result := w.Result()

		assert.Equal(t, tt.want.code, result.StatusCode)
		assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		assert.Equal(t, tt.want.location, result.Header.Get("Location"))

		err := result.Body.Close()
		require.NoError(t, err)
	}
}
