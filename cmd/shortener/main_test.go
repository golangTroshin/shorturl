package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/storage"
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
			name: "post_request_with_valid_body_response_valid_content_status_201",
			body: "https://practicum.yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
				content:     "/QrPnX5IU",
			},
		},
		{
			name: "post_request_with_empty_body_response_status_400",
			body: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				content:     "Empty body\n",
			},
		},
	}

	if err := config.ParseFlags(); err != nil {
		t.Fatalf("error ocured while parsing flags: %v", err)
	}

	for _, tt := range tests {
		store := storage.NewMemoryStore()
		router := Router(store)

		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
		r.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)
		result := w.Result()

		if tt.want.code != result.StatusCode {
			t.Errorf("[%s] codes are not equal: expected: %d, result: %d ", tt.name, tt.want.code, result.StatusCode)
		}

		if tt.want.contentType != result.Header.Get("Content-Type") {
			t.Errorf("[%s] content types are not equal: expected: %s, result: %s ", tt.name, tt.want.contentType, result.Header.Get("Content-Type"))
		}

		resultURL, err := io.ReadAll(result.Body)
		stingResultURL := string(resultURL)
		require.NoError(t, err)
		err = result.Body.Close()
		require.NoError(t, err)

		if tt.want.code == http.StatusCreated {
			expectedURL := config.Options.FlagBaseURL + tt.want.content
			if expectedURL != stingResultURL {
				t.Errorf("[%s] URLs are not equal: expected: %s, result: %s ", tt.name, expectedURL, stingResultURL)
			}
		} else {
			if tt.want.content != stingResultURL {
				t.Errorf("[%s] URLs are not equal: expected: %s, result: %s ", tt.name, tt.want.content, stingResultURL)
			}
		}
	}
}

func TestAPIPostHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		content     string
	}
	var request storage.RequestURL
	request.URL = "https://practicum.yandex.ru/"

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("error while Marshal result: %s", err.Error())
		return
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "api_post_request_with_valid_body_response_valid_content_status_201",
			body: string(body),
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
				content:     "/QrPnX5IU",
			},
		},
		{
			name: "api_post_request_with_empty_body_response_status_400",
			body: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				content:     "Wrong request body\n",
			},
		},
	}

	if err := config.ParseFlags(); err != nil {
		t.Fatalf("error ocured while parsing flags: %v", err)
	}

	var responseShortURL storage.ResponseShortURL

	for _, tt := range tests {
		store := storage.NewMemoryStore()
		router := Router(store)

		r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)
		result := w.Result()

		if tt.want.code != result.StatusCode {
			t.Errorf("[%s] codes are not equal: expected: %d, result: %d ", tt.name, tt.want.code, result.StatusCode)
		}

		if tt.want.contentType != result.Header.Get("Content-Type") {
			t.Errorf("[%s] content types are not equal: expected: %s, result: %s ", tt.name, tt.want.contentType, result.Header.Get("Content-Type"))
		}

		resultURL, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		if err = json.Unmarshal(resultURL, &responseShortURL); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = result.Body.Close()
		require.NoError(t, err)

		stringResultURL := responseShortURL.ShortURL
		if tt.want.code == http.StatusCreated {
			expectedURL := config.Options.FlagBaseURL + tt.want.content
			if expectedURL != stringResultURL {
				t.Errorf("[%s] URLs are not equal: expected: %s, result: %s ", tt.name, expectedURL, stringResultURL)
			}
		} else {
			if tt.want.content != stringResultURL {
				t.Errorf("[%s] URLs are not equal: expected: %s, result: %s ", tt.name, tt.want.content, stringResultURL)
			}
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
			name:       "get_request_existed_index_response_valid_location_header_status_307",
			requestURI: "/QrPnX5IU",
			want: want{
				code:        http.StatusTemporaryRedirect,
				contentType: "text/plain",
				location:    "https://practicum.yandex.ru/",
			},
		},
		{
			name:       "get_request_not_existed_index_response_empty_location_header_status_400",
			requestURI: "/aHR0cHM63444",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
		{
			name:       "get_request_not_existed_page_response_empty_location_header_status_404",
			requestURI: "/aHR0cHM6/123/qwewqe",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
	}

	if err := config.ParseFlags(); err != nil {
		t.Fatalf("error ocured while parsing flags: %v", err)
	}

	for _, tt := range tests {
		store := storage.NewMemoryStore()
		store.Set(context.Background(), "https://practicum.yandex.ru/")
		router := Router(store)

		r := httptest.NewRequest(http.MethodGet, tt.requestURI, nil)
		r.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)
		result := w.Result()

		if tt.want.code != result.StatusCode {
			t.Errorf("[%s] codes are not equal: expected: %d, result: %d ", tt.name, tt.want.code, result.StatusCode)
		}

		if tt.want.contentType != result.Header.Get("Content-Type") {
			t.Errorf("[%s] content types are not equal: expected: %s, result: %s ", tt.name, tt.want.contentType, result.Header.Get("Content-Type"))
		}

		if tt.want.location != result.Header.Get("Location") {
			t.Errorf("[%s] locations are not equal: expected: %s, result: %s ", tt.name, tt.want.location, result.Header.Get("Location"))
		}

		err := result.Body.Close()
		require.NoError(t, err)
	}
}
