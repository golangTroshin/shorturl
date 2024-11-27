package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golangTroshin/shorturl/internal/app/handlers"
	"github.com/golangTroshin/shorturl/internal/app/storage"
	"github.com/golangTroshin/shorturl/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAPIPostHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := handlers.APIPostHandler(mockStorage)

	originalURL := "https://example.com"
	shortenedURL := "EAaArVRs"

	mockStorage.EXPECT().Set(gomock.Any(), originalURL).Return(storage.URL{
		ShortURL: shortenedURL,
	}, nil)

	body := `{"url":"` + originalURL + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	responseBody := storage.ResponseShortURL{}
	err := json.NewDecoder(w.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "/"+shortenedURL, responseBody.ShortURL)
}

func TestAPIPostBatchHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	handler := handlers.APIPostBatchHandler(mockStorage)

	batch := []storage.RequestBodyBanch{
		{CorrelationID: "id1", OriginalURL: "https://example.com/1"},
		{CorrelationID: "id2", OriginalURL: "https://example.com/2"},
	}
	urlObjs := []storage.URL{
		{UUID: "id1", ShortURL: "8vl4QULk", OriginalURL: "https://example.com/1"},
		{UUID: "id2", ShortURL: "EFNP1MV_", OriginalURL: "https://example.com/2"},
	}

	mockStorage.EXPECT().SetBatch(gomock.Any(), batch).Return(urlObjs, nil)

	body, _ := json.Marshal(batch)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response []map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "/8vl4QULk", response[0]["short_url"])
	assert.Equal(t, "id1", response[0]["correlation_id"])
	assert.Equal(t, "/EFNP1MV_", response[1]["short_url"])
	assert.Equal(t, "id2", response[1]["correlation_id"])
}
