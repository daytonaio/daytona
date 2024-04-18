//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistries

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/containerregistry"
)

type InMemoryContainerRegistryStore struct {
	crs map[string]map[string]*containerregistry.ContainerRegistry
}

func NewInMemoryContainerRegistryStore() containerregistry.Store {
	return &InMemoryContainerRegistryStore{
		crs: make(map[string]map[string]*containerregistry.ContainerRegistry),
	}
}

func (s *InMemoryContainerRegistryStore) List() ([]*containerregistry.ContainerRegistry, error) {
	crs := []*containerregistry.ContainerRegistry{}
	//	itterate server key
	for _, s := range s.crs {
		//	itterate username key
		for _, u := range s {
			crs = append(crs, u)
		}
	}

	return crs, nil
}

func (s *InMemoryContainerRegistryStore) Find(server, username string) (*containerregistry.ContainerRegistry, error) {
	crs, ok := s.crs[server]
	if !ok {
		return nil, errors.New("container registry not found")
	}
	cru, ok := crs[username]
	if !ok {
		return nil, errors.New("container registry not found")
	}

	return cru, nil
}

func (s *InMemoryContainerRegistryStore) Save(cr *containerregistry.ContainerRegistry) error {
	_, ok := s.crs[cr.Server]
	if !ok {
		s.crs[cr.Server] = make(map[string]*containerregistry.ContainerRegistry)
	}
	s.crs[cr.Server][cr.Username] = cr
	return nil
}

func (s *InMemoryContainerRegistryStore) Delete(cr *containerregistry.ContainerRegistry) error {
	crs, ok := s.crs[cr.Server]
	if !ok {
		return errors.New("container registry not found")
	}
	_, ok = crs[cr.Username]
	if !ok {
		return errors.New("container registry not found")
	}
	delete(crs, cr.Username)
	s.crs[cr.Server] = crs
	return nil
}
