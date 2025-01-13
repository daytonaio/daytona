//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"io"

	"github.com/daytonaio/daytona/pkg/services"
	"github.com/stretchr/testify/mock"
)

type MockBuildService struct {
	mock.Mock
}

func NewMockBuildService() *MockBuildService {
	return &MockBuildService{}
}

func (m *MockBuildService) Create(createBuildDto services.CreateBuildDTO) (string, error) {
	args := m.Called(createBuildDto)
	return args.String(0), args.Error(1)
}

func (m *MockBuildService) Find(filter *services.BuildFilter) (*services.BuildDTO, error) {
	args := m.Called(filter)
	return args.Get(0).(*services.BuildDTO), args.Error(1)
}

func (m *MockBuildService) List(filter *services.BuildFilter) ([]*services.BuildDTO, error) {
	args := m.Called(filter)
	return args.Get(0).([]*services.BuildDTO), args.Error(1)
}

func (m *MockBuildService) Delete(filter *services.BuildFilter, force bool) []error {
	args := m.Called(filter, force)
	return args.Get(0).([]error)
}

func (m *MockBuildService) HandleSuccessfulRemoval(id string) error {
	args := m.Called(id)
	return args.Get(0).(error)
}

func (m *MockBuildService) GetBuildLogReader(buildId string) (io.Reader, error) {
	args := m.Called(buildId)
	return args.Get(0).(io.Reader), args.Error(1)
}
