//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/stretchr/testify/mock"
)

type mockContainerRegistryService struct {
	mock.Mock
}

func NewMockContainerRegistryService() *mockContainerRegistryService {
	return &mockContainerRegistryService{}
}

func (m *mockContainerRegistryService) Delete(server string) error {
	args := m.Called(server)
	return args.Error(0)
}

func (m *mockContainerRegistryService) Find(server string) (*containerregistry.ContainerRegistry, error) {
	args := m.Called(server)
	return args.Get(0).(*containerregistry.ContainerRegistry), args.Error(1)
}

func (m *mockContainerRegistryService) FindByImageName(imageName string) (*containerregistry.ContainerRegistry, error) {
	args := m.Called(imageName)
	return args.Get(0).(*containerregistry.ContainerRegistry), args.Error(1)
}

func (m *mockContainerRegistryService) List() ([]*containerregistry.ContainerRegistry, error) {
	args := m.Called()
	return args.Get(0).([]*containerregistry.ContainerRegistry), args.Error(1)
}

func (m *mockContainerRegistryService) Map() (map[string]*containerregistry.ContainerRegistry, error) {
	args := m.Called()
	return args.Get(0).(map[string]*containerregistry.ContainerRegistry), args.Error(1)
}

func (m *mockContainerRegistryService) Save(cr *containerregistry.ContainerRegistry) error {
	args := m.Called(cr)
	return args.Error(0)
}
