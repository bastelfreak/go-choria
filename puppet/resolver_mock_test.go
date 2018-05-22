// Code generated by MockGen. DO NOT EDIT.
// Source: puppet_security.go

// Package provider is a generated GoMock package.
package provider

import (
	srvcache "github.com/choria-io/go-choria/srvcache"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockResolver is a mock of Resolver interface
type MockResolver struct {
	ctrl     *gomock.Controller
	recorder *MockResolverMockRecorder
}

// MockResolverMockRecorder is the mock recorder for MockResolver
type MockResolverMockRecorder struct {
	mock *MockResolver
}

// NewMockResolver creates a new mock instance
func NewMockResolver(ctrl *gomock.Controller) *MockResolver {
	mock := &MockResolver{ctrl: ctrl}
	mock.recorder = &MockResolverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockResolver) EXPECT() *MockResolverMockRecorder {
	return m.recorder
}

// QuerySrvRecords mocks base method
func (m *MockResolver) QuerySrvRecords(records []string) ([]srvcache.Server, error) {
	ret := m.ctrl.Call(m, "QuerySrvRecords", records)
	ret0, _ := ret[0].([]srvcache.Server)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QuerySrvRecords indicates an expected call of QuerySrvRecords
func (mr *MockResolverMockRecorder) QuerySrvRecords(records interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QuerySrvRecords", reflect.TypeOf((*MockResolver)(nil).QuerySrvRecords), records)
}
