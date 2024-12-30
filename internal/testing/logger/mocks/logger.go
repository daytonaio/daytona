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

func (f *MockLoggerFactory) CreateWorkspaceLogger(workspaceId string, source logs.LogSource) logs.Logger {
	return &mockLogger{}
}

func (f *MockLoggerFactory) CreateProjectLogger(workspaceId, projectName string, source logs.LogSource) logs.Logger {
	return &mockLogger{}
}

func (f *MockLoggerFactory) CreateBuildLogger(projectName, hash string, source logs.LogSource) logs.Logger {
	return &mockLogger{}
}

func (f *MockLoggerFactory) CreateWorkspaceLogReader(workspaceId string) (io.Reader, error) {
	return nil, nil
}

func (f *MockLoggerFactory) CreateProjectLogReader(workspaceId, projectName string) (io.Reader, error) {
	return nil, nil
}

func (f *MockLoggerFactory) CreateBuildLogReader(projectName, hash string) (io.Reader, error) {
	return nil, nil
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
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
