//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockScheduler struct {
	mock.Mock
}

func (s *MockScheduler) Start() {
	s.Called()
}

func (s *MockScheduler) Stop() {
	s.Called()
}

func (s *MockScheduler) AddFunc(spec string, cmd func()) error {
	args := s.Called(spec, cmd)
	return args.Error(0)
}
