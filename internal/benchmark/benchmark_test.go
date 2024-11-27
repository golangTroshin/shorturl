package benchmark_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golangTroshin/shorturl/internal/app/handlers"
	"github.com/golangTroshin/shorturl/internal/app/helpers"
	"github.com/golangTroshin/shorturl/internal/app/middleware"
	"github.com/golangTroshin/shorturl/internal/app/storage"
	"github.com/golangTroshin/shorturl/internal/mocks"
)

func BenchmarkUserIDGenerator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		helpers.GenerateRandomUserID(10)
	}
}

func BenchmarkBuildJWTString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		helpers.BuildJWTString()
	}
}

func BenchmarkGetUserIDByToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		helpers.GetUserIDByToken(helpers.GenerateRandomUserID(10))
	}
}

func BenchmarkAPIPostHandler(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorage(ctrl)
	handler := handlers.APIPostHandler(mockStore)

	mockStore.EXPECT().
		Set(gomock.Any(), "https://example.com").
		Return(storage.URL{ShortURL: "EAaArVRs"}, nil).
		AnyTimes()

	body := storage.RequestURL{URL: "https://example.com"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkAPIPostBatchHandler(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorage(ctrl)
	handler := handlers.APIPostBatchHandler(mockStore)

	requestBodies := []storage.RequestBodyBanch{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
		{CorrelationID: "2", OriginalURL: "https://another.com"},
	}
	urlObjs := []storage.URL{
		{UUID: "1", ShortURL: "EAaArVRs"},
		{UUID: "2", ShortURL: "BMOSbMDJ"},
	}

	mockStore.EXPECT().
		SetBatch(gomock.Any(), requestBodies).
		Return(urlObjs, nil).
		AnyTimes()

	bodyBytes, _ := json.Marshal(requestBodies)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkAPIDeleteUrlsHandler(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorage(ctrl)
	handler := handlers.APIDeleteUrlsHandler(mockStore)

	urlIDs := []string{"short1", "short2"}
	userID := "user1"

	mockStore.EXPECT().
		BatchDeleteURLs(userID, urlIDs).
		Return(nil).
		AnyTimes()

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)

	bodyBytes, _ := json.Marshal(urlIDs)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(bodyBytes))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkPostRequestHandler(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorage(ctrl)
	handler := handlers.PostRequestHandler(mockStore)

	mockStore.EXPECT().
		Set(gomock.Any(), "https://example.com").
		Return(storage.URL{ShortURL: "EAaArVRs"}, nil).
		AnyTimes()

	body := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "text/plain")

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkGetRequestHandler(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorage(ctrl)
	handler := handlers.GetRequestHandler(mockStore)

	mockStore.EXPECT().
		Get(gomock.Any(), "EAaArVRs").
		Return("https://example.com", nil).
		AnyTimes()

	req := httptest.NewRequest(http.MethodGet, "/EAaArVRs", nil)
	req.Header.Set("Content-Type", "text/plain")

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkGetURLsByUserHandler(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorage(ctrl)
	handler := handlers.GetURLsByUserHandler(mockStore)

	mockStore.EXPECT().
		GetByUserID(gomock.Any(), "user123").
		Return([]storage.URL{
			{ShortURL: "abc123", OriginalURL: "https://example.com"},
			{ShortURL: "xyz789", OriginalURL: "https://another.com"},
		}, nil).
		AnyTimes()

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "user123")
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
