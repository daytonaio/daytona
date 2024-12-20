//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/stretchr/testify/mock"
)

type MockLoggerPlugin struct {
	mock.Mock
}

type MockLoggerFactory struct {
	mock.Mock
}

func (f *MockLoggerFactory) CreateTargetLogger(targetId string, source logs.LogSource) logs.Logger {
	return &mockLogger{}
}

func (f *MockLoggerFactory) CreateWorkspaceLogger(targetId, workspaceName string, source logs.LogSource) logs.Logger {
	return &mockLogger{}
}

func (f *MockLoggerFactory) CreateBuildLogger(workspaceName, hash string, source logs.LogSource) logs.Logger {
	return &mockLogger{}
}

func (f *MockLoggerFactory) CreateTargetLogReader(targetId string) (io.Reader, error) {
	return nil, nil
}

func (f *MockLoggerFactory) CreateWorkspaceLogReader(targetId, workspaceName string) (io.Reader, error) {
	return nil, nil
}

func (f *MockLoggerFactory) CreateBuildLogReader(workspaceName, hash string) (io.Reader, error) {
	return nil, nil
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockLogger) ConstructJsonLogEntry(p []byte) ([]byte, error) {
	args := m.Called(p)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockLogger) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockLogger) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func NewMockLogger() *mockLogger {
	return &mockLogger{}
}
