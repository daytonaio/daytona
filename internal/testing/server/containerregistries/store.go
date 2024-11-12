//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistries

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
)

type InMemoryContainerRegistryStore struct {
	crs map[string]*models.ContainerRegistry
}

func NewInMemoryContainerRegistryStore() containerregistries.ContainerRegistryStore {
	return &InMemoryContainerRegistryStore{
		crs: make(map[string]*models.ContainerRegistry),
	}
}

func (s *InMemoryContainerRegistryStore) List() ([]*models.ContainerRegistry, error) {
	crs := []*models.ContainerRegistry{}
	for _, cr := range s.crs {
		crs = append(crs, cr)
	}

	return crs, nil
}

func (s *InMemoryContainerRegistryStore) Find(server string) (*models.ContainerRegistry, error) {
	cr, ok := s.crs[server]
	if !ok {
		return nil, containerregistries.ErrContainerRegistryNotFound
	}

	return cr, nil
}

func (s *InMemoryContainerRegistryStore) Save(cr *models.ContainerRegistry) error {
	s.crs[cr.Server] = cr
	return nil
}

func (s *InMemoryContainerRegistryStore) Delete(cr *models.ContainerRegistry) error {
	_, ok := s.crs[cr.Server]
	if !ok {
		return containerregistries.ErrContainerRegistryNotFound
	}
	delete(s.crs, cr.Server)
	return nil
}
