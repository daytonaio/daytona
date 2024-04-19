// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistries

import "github.com/daytonaio/daytona/pkg/containerregistry"

type IContainerRegistryService interface {
	Delete(server string) error
	Find(server string) (*containerregistry.ContainerRegistry, error)
	List() ([]*containerregistry.ContainerRegistry, error)
	Map() (map[string]*containerregistry.ContainerRegistry, error)
	Save(cr *containerregistry.ContainerRegistry) error
}

type ContainerRegistryServiceConfig struct {
	Store containerregistry.Store
}

type ContainerRegistryService struct {
	store containerregistry.Store
}

func NewContainerRegistryService(config ContainerRegistryServiceConfig) IContainerRegistryService {
	return &ContainerRegistryService{
		store: config.Store,
	}
}

func (s *ContainerRegistryService) List() ([]*containerregistry.ContainerRegistry, error) {
	return s.store.List()
}

func (s *ContainerRegistryService) Map() (map[string]*containerregistry.ContainerRegistry, error) {
	list, err := s.store.List()
	if err != nil {
		return nil, err
	}

	crs := make(map[string]*containerregistry.ContainerRegistry)
	for _, cr := range list {
		crs[cr.Server] = cr
	}

	return crs, nil
}

func (s *ContainerRegistryService) Find(server string) (*containerregistry.ContainerRegistry, error) {
	return s.store.Find(server)
}

func (s *ContainerRegistryService) Save(cr *containerregistry.ContainerRegistry) error {
	return s.store.Save(cr)
}

func (s *ContainerRegistryService) Delete(server string) error {
	cr, err := s.Find(server)
	if err != nil {
		return err
	}
	return s.store.Delete(cr)
}
