//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/stretchr/testify/mock"
)

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
