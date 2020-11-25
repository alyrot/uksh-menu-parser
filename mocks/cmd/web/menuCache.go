// Code generated by MockGen. DO NOT EDIT.
// Source: menuCache.go

// Package mock_main is a generated GoMock package.
package mock_main

import (
	parser "github.com/alyrot/uksh-menu-parser/pkg/parser"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockMenuCacher is a mock of MenuCacher interface
type MockMenuCacher struct {
	ctrl     *gomock.Controller
	recorder *MockMenuCacherMockRecorder
}

// MockMenuCacherMockRecorder is the mock recorder for MockMenuCacher
type MockMenuCacherMockRecorder struct {
	mock *MockMenuCacher
}

// NewMockMenuCacher creates a new mock instance
func NewMockMenuCacher(ctrl *gomock.Controller) *MockMenuCacher {
	mock := &MockMenuCacher{ctrl: ctrl}
	mock.recorder = &MockMenuCacherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMenuCacher) EXPECT() *MockMenuCacherMockRecorder {
	return m.recorder
}

// GetMenu mocks base method
func (m *MockMenuCacher) GetMenu(date time.Time) ([]*parser.Dish, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMenu", date)
	ret0, _ := ret[0].([]*parser.Dish)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMenu indicates an expected call of GetMenu
func (mr *MockMenuCacherMockRecorder) GetMenu(date interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMenu", reflect.TypeOf((*MockMenuCacher)(nil).GetMenu), date)
}

// Refresh mocks base method
func (m *MockMenuCacher) Refresh() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refresh")
	ret0, _ := ret[0].(error)
	return ret0
}

// Refresh indicates an expected call of Refresh
func (mr *MockMenuCacherMockRecorder) Refresh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refresh", reflect.TypeOf((*MockMenuCacher)(nil).Refresh))
}

// MockDownloader is a mock of Downloader interface
type MockDownloader struct {
	ctrl     *gomock.Controller
	recorder *MockDownloaderMockRecorder
}

// MockDownloaderMockRecorder is the mock recorder for MockDownloader
type MockDownloaderMockRecorder struct {
	mock *MockDownloader
}

// NewMockDownloader creates a new mock instance
func NewMockDownloader(ctrl *gomock.Controller) *MockDownloader {
	mock := &MockDownloader{ctrl: ctrl}
	mock.recorder = &MockDownloaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDownloader) EXPECT() *MockDownloaderMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockDownloader) Get(url string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", url)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockDownloaderMockRecorder) Get(url interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDownloader)(nil).Get), url)
}
