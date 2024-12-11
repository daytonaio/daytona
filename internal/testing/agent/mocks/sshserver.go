//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"errors"

	"github.com/stretchr/testify/mock"
)

var SshServerStartError = errors.New("start error")

type mockSshServer struct {
	mock.Mock
}

func (m *mockSshServer) Start() error {
	args := m.Called()
	return args.Error(0)
}

func NewMockSshServer() *mockSshServer {
	mockSshServer := new(mockSshServer)
	mockSshServer.On("Start").Return(SshServerStartError)

	return mockSshServer
}
