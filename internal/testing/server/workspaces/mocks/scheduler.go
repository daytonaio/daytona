//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockSchedulerPlugin struct {
	mock.Mock
}

func (s *MockSchedulerPlugin) Start() {
	s.Called()
}

func (s *MockSchedulerPlugin) Stop() {
	s.Called()
}

func (s *MockSchedulerPlugin) AddFunc(spec string, cmd func()) error {
	args := s.Called(spec, cmd)
	return args.Error(0)
}
