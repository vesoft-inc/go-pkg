// Code generated by MockGen. DO NOT EDIT.
// Source: mail.go

// Package mail is a generated GoMock package.
package mail

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	gomail "gopkg.in/gomail.v2"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// Send mocks base method.
func (m *MockClient) Send(to []string, subject, contentType, body string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", to, subject, contentType, body)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockClientMockRecorder) Send(to, subject, contentType, body interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockClient)(nil).Send), to, subject, contentType, body)
}

// MockdialerSender is a mock of dialerSender interface.
type MockdialerSender struct {
	ctrl     *gomock.Controller
	recorder *MockdialerSenderMockRecorder
}

// MockdialerSenderMockRecorder is the mock recorder for MockdialerSender.
type MockdialerSenderMockRecorder struct {
	mock *MockdialerSender
}

// NewMockdialerSender creates a new mock instance.
func NewMockdialerSender(ctrl *gomock.Controller) *MockdialerSender {
	mock := &MockdialerSender{ctrl: ctrl}
	mock.recorder = &MockdialerSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockdialerSender) EXPECT() *MockdialerSenderMockRecorder {
	return m.recorder
}

// DialAndSend mocks base method.
func (m_2 *MockdialerSender) DialAndSend(m ...*gomail.Message) error {
	m_2.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range m {
		varargs = append(varargs, a)
	}
	ret := m_2.ctrl.Call(m_2, "DialAndSend", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DialAndSend indicates an expected call of DialAndSend.
func (mr *MockdialerSenderMockRecorder) DialAndSend(m ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DialAndSend", reflect.TypeOf((*MockdialerSender)(nil).DialAndSend), m...)
}
