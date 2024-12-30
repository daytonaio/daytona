//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/stretchr/testify/mock"
)

type mockToolboxServer struct {
	mock.Mock
}

func (m *mockToolboxServer) Start() error {
	args := m.Called()
	return args.Error(0)
}

func NewMockToolboxServer() *mockToolboxServer {
	mockToolboxServer := new(mockToolboxServer)
	mockToolboxServer.On("Start").Return(nil)

	return mockToolboxServer
}
