// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

type Store interface {
	List() ([]*ContainerRegistry, error)
	Find(server string, username string) (*ContainerRegistry, error)
	Save(cr *ContainerRegistry) error
	Delete(cr *ContainerRegistry) error
}
