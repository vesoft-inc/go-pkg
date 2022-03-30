// Code generated by MockGen. DO NOT EDIT.
// Source: notifier.go

// Package notify is a generated GoMock package.
package notify

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockNotifier is a mock of Notifier interface.
type MockNotifier struct {
	ctrl     *gomock.Controller
	recorder *MockNotifierMockRecorder
}

// MockNotifierMockRecorder is the mock recorder for MockNotifier.
type MockNotifierMockRecorder struct {
	mock *MockNotifier
}

// NewMockNotifier creates a new mock instance.
func NewMockNotifier(ctrl *gomock.Controller) *MockNotifier {
	mock := &MockNotifier{ctrl: ctrl}
	mock.recorder = &MockNotifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNotifier) EXPECT() *MockNotifierMockRecorder {
	return m.recorder
}

// Notify mocks base method.
func (m *MockNotifier) Notify(arg0 context.Context, arg1 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Notify", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Notify indicates an expected call of Notify.
func (mr *MockNotifierMockRecorder) Notify(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notify", reflect.TypeOf((*MockNotifier)(nil).Notify), arg0, arg1)
}

// MockStringNotifier is a mock of StringNotifier interface.
type MockStringNotifier struct {
	ctrl     *gomock.Controller
	recorder *MockStringNotifierMockRecorder
}

// MockStringNotifierMockRecorder is the mock recorder for MockStringNotifier.
type MockStringNotifierMockRecorder struct {
	mock *MockStringNotifier
}

// NewMockStringNotifier creates a new mock instance.
func NewMockStringNotifier(ctrl *gomock.Controller) *MockStringNotifier {
	mock := &MockStringNotifier{ctrl: ctrl}
	mock.recorder = &MockStringNotifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStringNotifier) EXPECT() *MockStringNotifierMockRecorder {
	return m.recorder
}

// Notify mocks base method.
func (m *MockStringNotifier) Notify(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Notify", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Notify indicates an expected call of Notify.
func (mr *MockStringNotifierMockRecorder) Notify(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notify", reflect.TypeOf((*MockStringNotifier)(nil).Notify), arg0, arg1)
}
