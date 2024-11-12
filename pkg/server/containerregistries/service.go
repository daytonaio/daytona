// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistries

import (
	"strings"

	"github.com/daytonaio/daytona/pkg/models"
)

type IContainerRegistryService interface {
	Delete(server string) error
	Find(server string) (*models.ContainerRegistry, error)
	FindByImageName(imageName string) (*models.ContainerRegistry, error)
	List() ([]*models.ContainerRegistry, error)
	Map() (map[string]*models.ContainerRegistry, error)
	Save(cr *models.ContainerRegistry) error
}

type ContainerRegistryServiceConfig struct {
	Store ContainerRegistryStore
}

type ContainerRegistryService struct {
	store ContainerRegistryStore
}

func NewContainerRegistryService(config ContainerRegistryServiceConfig) IContainerRegistryService {
	return &ContainerRegistryService{
		store: config.Store,
	}
}

func (s *ContainerRegistryService) List() ([]*models.ContainerRegistry, error) {
	return s.store.List()
}

func (s *ContainerRegistryService) Map() (map[string]*models.ContainerRegistry, error) {
	list, err := s.store.List()
	if err != nil {
		return nil, err
	}

	crs := make(map[string]*models.ContainerRegistry)
	for _, cr := range list {
		crs[cr.Server] = cr
	}

	return crs, nil
}

func (s *ContainerRegistryService) Find(server string) (*models.ContainerRegistry, error) {
	return s.store.Find(server)
}

func (s *ContainerRegistryService) FindByImageName(imageName string) (*models.ContainerRegistry, error) {
	server := getImageServer(imageName)

	return s.Find(server)
}

func (s *ContainerRegistryService) Save(cr *models.ContainerRegistry) error {
	return s.store.Save(cr)
}

func (s *ContainerRegistryService) Delete(server string) error {
	cr, err := s.Find(server)
	if err != nil {
		return err
	}
	return s.store.Delete(cr)
}

func getImageServer(imageName string) string {
	parts := strings.Split(imageName, "/")

	if len(parts) < 3 {
		return "docker.io"
	}

	return parts[0]
}
