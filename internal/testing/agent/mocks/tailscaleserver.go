//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import "github.com/stretchr/testify/mock"

type mockTailscaleServer struct {
	mock.Mock
}

func (m *mockTailscaleServer) Start() error {
	args := m.Called()
	return args.Error(0)
}

func NewMockTailscaleServer() *mockTailscaleServer {
	mockTailscaleServer := new(mockTailscaleServer)
	mockTailscaleServer.On("Start").Return(nil)

	return mockTailscaleServer
}
