// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
)

type ContainerRegistryStore struct {
	db *gorm.DB
}

func NewContainerRegistryStore(db *gorm.DB) (*ContainerRegistryStore, error) {
	err := db.AutoMigrate(&models.ContainerRegistry{})
	if err != nil {
		return nil, err
	}

	return &ContainerRegistryStore{db: db}, nil
}

func (s *ContainerRegistryStore) List() ([]*models.ContainerRegistry, error) {
	containerRegistries := []*models.ContainerRegistry{}
	tx := s.db.Find(&containerRegistries)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return containerRegistries, nil
}

func (s *ContainerRegistryStore) Find(server string) (*models.ContainerRegistry, error) {
	containerRegistry := &models.ContainerRegistry{}
	tx := s.db.Where("server = ?", server).First(containerRegistry)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, containerregistries.ErrContainerRegistryNotFound
		}
		return nil, tx.Error
	}

	return containerRegistry, nil
}

func (s *ContainerRegistryStore) Save(cr *models.ContainerRegistry) error {
	tx := s.db.Save(cr)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *ContainerRegistryStore) Delete(cr *models.ContainerRegistry) error {
	tx := s.db.Delete(cr)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return containerregistries.ErrContainerRegistryNotFound
	}

	return nil
}
