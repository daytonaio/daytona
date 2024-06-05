//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistries

import (
	"github.com/daytonaio/daytona/pkg/containerregistry"
)

type InMemoryContainerRegistryStore struct {
	crs map[string]*containerregistry.ContainerRegistry
}

func NewInMemoryContainerRegistryStore() containerregistry.Store {
	return &InMemoryContainerRegistryStore{
		crs: make(map[string]*containerregistry.ContainerRegistry),
	}
}

func (s *InMemoryContainerRegistryStore) List() ([]*containerregistry.ContainerRegistry, error) {
	crs := []*containerregistry.ContainerRegistry{}
	for _, cr := range s.crs {
		crs = append(crs, cr)
	}

	return crs, nil
}

func (s *InMemoryContainerRegistryStore) Find(server string) (*containerregistry.ContainerRegistry, error) {
	cr, ok := s.crs[server]
	if !ok {
		return nil, containerregistry.ErrContainerRegistryNotFound
	}

	return cr, nil
}

func (s *InMemoryContainerRegistryStore) Save(cr *containerregistry.ContainerRegistry) error {
	s.crs[cr.Server] = cr
	return nil
}

func (s *InMemoryContainerRegistryStore) Delete(cr *containerregistry.ContainerRegistry) error {
	_, ok := s.crs[cr.Server]
	if !ok {
		return containerregistry.ErrContainerRegistryNotFound
	}
	delete(s.crs, cr.Server)
	return nil
}
