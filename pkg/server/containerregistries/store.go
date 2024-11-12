// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistries

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type ContainerRegistryStore interface {
	List() ([]*models.ContainerRegistry, error)
	Find(server string) (*models.ContainerRegistry, error)
	Save(cr *models.ContainerRegistry) error
	Delete(cr *models.ContainerRegistry) error
}

var (
	ErrContainerRegistryNotFound = errors.New("container registry not found")
)

func IsContainerRegistryNotFound(err error) bool {
	return err.Error() == ErrContainerRegistryNotFound.Error()
}
