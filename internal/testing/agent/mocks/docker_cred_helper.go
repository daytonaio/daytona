//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import "github.com/stretchr/testify/mock"

type mockDockerCredHelper struct {
	mock.Mock
}

func (m *mockDockerCredHelper) SetDockerConfig() error {
	args := m.Called()
	return args.Error(0)
}

func NewMockDockerCredHelper() *mockDockerCredHelper {
	mockCredHelper := new(mockDockerCredHelper)
	mockCredHelper.On("SetDockerConfig").Return(nil)

	return mockCredHelper
}
