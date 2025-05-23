// Code generated by MockGen. DO NOT EDIT.
// Source: hash.go
//
// Generated by this command:
//
//	mockgen -source=hash.go -destination=./hash_mock.go -package=hash
//

// Package hash is a generated GoMock package.
package hash

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockKeccakState is a mock of KeccakState interface.
type MockKeccakState struct {
	ctrl     *gomock.Controller
	recorder *MockKeccakStateMockRecorder
	isgomock struct{}
}

// MockKeccakStateMockRecorder is the mock recorder for MockKeccakState.
type MockKeccakStateMockRecorder struct {
	mock *MockKeccakState
}

// NewMockKeccakState creates a new mock instance.
func NewMockKeccakState(ctrl *gomock.Controller) *MockKeccakState {
	mock := &MockKeccakState{ctrl: ctrl}
	mock.recorder = &MockKeccakStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeccakState) EXPECT() *MockKeccakStateMockRecorder {
	return m.recorder
}

// BlockSize mocks base method.
func (m *MockKeccakState) BlockSize() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BlockSize")
	ret0, _ := ret[0].(int)
	return ret0
}

// BlockSize indicates an expected call of BlockSize.
func (mr *MockKeccakStateMockRecorder) BlockSize() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BlockSize", reflect.TypeOf((*MockKeccakState)(nil).BlockSize))
}

// Read mocks base method.
func (m *MockKeccakState) Read(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockKeccakStateMockRecorder) Read(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockKeccakState)(nil).Read), arg0)
}

// Reset mocks base method.
func (m *MockKeccakState) Reset() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Reset")
}

// Reset indicates an expected call of Reset.
func (mr *MockKeccakStateMockRecorder) Reset() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reset", reflect.TypeOf((*MockKeccakState)(nil).Reset))
}

// Size mocks base method.
func (m *MockKeccakState) Size() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(int)
	return ret0
}

// Size indicates an expected call of Size.
func (mr *MockKeccakStateMockRecorder) Size() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockKeccakState)(nil).Size))
}

// Sum mocks base method.
func (m *MockKeccakState) Sum(b []byte) []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sum", b)
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Sum indicates an expected call of Sum.
func (mr *MockKeccakStateMockRecorder) Sum(b any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sum", reflect.TypeOf((*MockKeccakState)(nil).Sum), b)
}

// Write mocks base method.
func (m *MockKeccakState) Write(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockKeccakStateMockRecorder) Write(p any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockKeccakState)(nil).Write), p)
}
