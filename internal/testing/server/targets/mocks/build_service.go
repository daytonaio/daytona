//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"io"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
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

func (m *MockBuildService) Find(filter *stores.BuildFilter) (*models.Build, error) {
	args := m.Called(filter)
	return args.Get(0).(*models.Build), args.Error(1)
}

func (m *MockBuildService) List(filter *stores.BuildFilter) ([]*models.Build, error) {
	args := m.Called(filter)
	return args.Get(0).([]*models.Build), args.Error(1)
}

func (m *MockBuildService) MarkForDeletion(filter *stores.BuildFilter, force bool) []error {
	args := m.Called(filter, force)
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
