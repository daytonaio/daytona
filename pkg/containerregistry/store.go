// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import "errors"

type Store interface {
	List() ([]*ContainerRegistry, error)
	Find(server string) (*ContainerRegistry, error)
	Save(cr *ContainerRegistry) error
	Delete(cr *ContainerRegistry) error
}

var (
	ErrContainerRegistryNotFound = errors.New("container registry not found")
)

func IsContainerRegistryNotFound(err error) bool {
	return err.Error() == ErrContainerRegistryNotFound.Error()
}
