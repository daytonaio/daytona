//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"io"
	"time"

	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/server/builds/dto"
	"github.com/stretchr/testify/mock"
)

type MockBuildService struct {
	mock.Mock
}

func NewMockBuildService() *MockBuildService {
	return &MockBuildService{}
}

func (m *MockBuildService) Create(createBuildDto dto.BuildCreationData) (string, error) {
	args := m.Called(createBuildDto)
	return args.String(0), args.Error(1)
}

func (m *MockBuildService) Find(filter *build.Filter) (*build.Build, error) {
	args := m.Called(filter)
	return args.Get(0).(*build.Build), args.Error(1)
}

func (m *MockBuildService) List(filter *build.Filter) ([]*build.Build, error) {
	args := m.Called(filter)
	return args.Get(0).([]*build.Build), args.Error(1)
}

func (m *MockBuildService) MarkForDeletion(filter *build.Filter) []error {
	args := m.Called(filter)
	return args.Get(0).([]error)
}

func (m *MockBuildService) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBuildService) AwaitEmptyList(waitTime time.Duration) error {
	args := m.Called(waitTime)
	return args.Error(0)
}

func (m *MockBuildService) GetBuildLogReader(buildId string) (io.Reader, error) {
	args := m.Called(buildId)
	return args.Get(0).(io.Reader), args.Error(1)
}
