//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/stretchr/testify/mock"
)

type MockBuildService struct {
	mock.Mock
}

func NewMockBuildService() *MockBuildService {
	return &MockBuildService{}
}

func (m *MockBuildService) Create(b *models.Build) error {
	args := m.Called(b)
	return args.Error(0)
}

func (m *MockBuildService) Update(b *models.Build) error {
	args := m.Called(b)
	return args.Error(0)
}

func (m *MockBuildService) Find(id string) (*models.Build, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Build), args.Error(1)
}

func (m *MockBuildService) List(filter *stores.BuildFilter) ([]*models.Build, error) {
	args := m.Called(filter)
	return args.Get(0).([]*models.Build), args.Error(1)
}

func (m *MockBuildService) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
