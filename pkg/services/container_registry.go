// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import "github.com/daytonaio/daytona/pkg/models"

type IContainerRegistryService interface {
	Delete(server string) error
	Find(server string) (*models.ContainerRegistry, error)
	FindByImageName(imageName string) (*models.ContainerRegistry, error)
	List() ([]*models.ContainerRegistry, error)
	Map() (map[string]*models.ContainerRegistry, error)
	Save(cr *models.ContainerRegistry) error
}
