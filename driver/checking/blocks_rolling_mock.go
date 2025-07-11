// Code generated by MockGen. DO NOT EDIT.
// Source: blocks_rolling.go
//
// Generated by this command:
//
//	mockgen -source blocks_rolling.go -destination blocks_rolling_mock.go -package checking
//
// Package checking is a generated GoMock package.
package checking

import (
	reflect "reflect"

	monitoring "github.com/0xsoniclabs/hyperion/driver/monitoring"
	gomock "go.uber.org/mock/gomock"
)

// MockMonitoringData is a mock of MonitoringData interface.
type MockMonitoringData struct {
	ctrl     *gomock.Controller
	recorder *MockMonitoringDataMockRecorder
}

// MockMonitoringDataMockRecorder is the mock recorder for MockMonitoringData.
type MockMonitoringDataMockRecorder struct {
	mock *MockMonitoringData
}

// NewMockMonitoringData creates a new mock instance.
func NewMockMonitoringData(ctrl *gomock.Controller) *MockMonitoringData {
	mock := &MockMonitoringData{ctrl: ctrl}
	mock.recorder = &MockMonitoringDataMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMonitoringData) EXPECT() *MockMonitoringDataMockRecorder {
	return m.recorder
}

// GetData mocks base method.
func (m *MockMonitoringData) GetData(arg0 monitoring.Node) monitoring.Series[monitoring.Time, monitoring.BlockStatus] {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetData", arg0)
	ret0, _ := ret[0].(monitoring.Series[monitoring.Time, monitoring.BlockStatus])
	return ret0
}

// GetData indicates an expected call of GetData.
func (mr *MockMonitoringDataMockRecorder) GetData(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetData", reflect.TypeOf((*MockMonitoringData)(nil).GetData), arg0)
}

// GetNodes mocks base method.
func (m *MockMonitoringData) GetNodes() []monitoring.Node {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodes")
	ret0, _ := ret[0].([]monitoring.Node)
	return ret0
}

// GetNodes indicates an expected call of GetNodes.
func (mr *MockMonitoringDataMockRecorder) GetNodes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodes", reflect.TypeOf((*MockMonitoringData)(nil).GetNodes))
}
