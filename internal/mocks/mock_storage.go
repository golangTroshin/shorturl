// Code generated by MockGen. DO NOT EDIT.
// Source: shorturl/internal/app/storage/storage.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	storage "github.com/golangTroshin/shorturl/internal/app/storage"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// BatchDeleteURLs mocks base method.
func (m *MockStorage) BatchDeleteURLs(userID string, batch []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchDeleteURLs", userID, batch)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchDeleteURLs indicates an expected call of BatchDeleteURLs.
func (mr *MockStorageMockRecorder) BatchDeleteURLs(userID, batch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchDeleteURLs", reflect.TypeOf((*MockStorage)(nil).BatchDeleteURLs), userID, batch)
}

// Get mocks base method.
func (m *MockStorage) Get(ctx context.Context, key string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, key)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStorageMockRecorder) Get(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStorage)(nil).Get), ctx, key)
}

// GetByUserID mocks base method.
func (m *MockStorage) GetByUserID(ctx context.Context, userID string) ([]storage.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByUserID", ctx, userID)
	ret0, _ := ret[0].([]storage.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByUserID indicates an expected call of GetByUserID.
func (mr *MockStorageMockRecorder) GetByUserID(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByUserID", reflect.TypeOf((*MockStorage)(nil).GetByUserID), ctx, userID)
}

// GetStats mocks base method.
func (m *MockStorage) GetStats(ctx context.Context) (storage.Stats, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStats", ctx)
	ret0, _ := ret[0].(storage.Stats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStats indicates an expected call of GetStats.
func (mr *MockStorageMockRecorder) GetStats(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStats", reflect.TypeOf((*MockStorage)(nil).GetStats), ctx)
}

// Set mocks base method.
func (m *MockStorage) Set(ctx context.Context, value string) (storage.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, value)
	ret0, _ := ret[0].(storage.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Set indicates an expected call of Set.
func (mr *MockStorageMockRecorder) Set(ctx, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockStorage)(nil).Set), ctx, value)
}

// SetBatch mocks base method.
func (m *MockStorage) SetBatch(ctx context.Context, batch []storage.RequestBodyBanch) ([]storage.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetBatch", ctx, batch)
	ret0, _ := ret[0].([]storage.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetBatch indicates an expected call of SetBatch.
func (mr *MockStorageMockRecorder) SetBatch(ctx, batch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBatch", reflect.TypeOf((*MockStorage)(nil).SetBatch), ctx, batch)
}
