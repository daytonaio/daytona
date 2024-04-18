// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	. "github.com/daytonaio/daytona/pkg/db/dto"
)

type ContainerRegistryStore struct {
	db *gorm.DB
}

func NewContainerRegistryStore(db *gorm.DB) (*ContainerRegistryStore, error) {
	err := db.AutoMigrate(&ContainerRegistryDTO{})
	if err != nil {
		return nil, err
	}

	return &ContainerRegistryStore{db: db}, nil
}

func (s *ContainerRegistryStore) List() ([]*containerregistry.ContainerRegistry, error) {
	containerRegistryDTOs := []ContainerRegistryDTO{}
	tx := s.db.Find(&containerRegistryDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	containerregistryTargets := []*containerregistry.ContainerRegistry{}
	for _, containerRegistryDTO := range containerRegistryDTOs {
		containerregistryTargets = append(containerregistryTargets, ToContainerRegistry(containerRegistryDTO))
	}

	return containerregistryTargets, nil
}

func (s *ContainerRegistryStore) Find(server string) (*containerregistry.ContainerRegistry, error) {
	containerRegistryDTO := ContainerRegistryDTO{}
	tx := s.db.Where("server = ?", server).First(&containerRegistryDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return ToContainerRegistry(containerRegistryDTO), nil
}

func (s *ContainerRegistryStore) Save(cr *containerregistry.ContainerRegistry) error {
	tx := s.db.Save(ToContainerRegistryDTO(cr))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *ContainerRegistryStore) Delete(cr *containerregistry.ContainerRegistry) error {
	tx := s.db.Delete(ToContainerRegistryDTO(cr))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
