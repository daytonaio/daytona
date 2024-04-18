// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/containerregistry"
)

type ContainerRegistryDTO struct {
	Server   string `gorm:"primaryKey"`
	Username string `gorm:"primaryKey"`
	Password string `json:"password"`
}

func ToContainerRegistryDTO(cr *containerregistry.ContainerRegistry) ContainerRegistryDTO {
	dto := ContainerRegistryDTO{
		Server:   cr.Server,
		Username: cr.Username,
		Password: cr.Password,
	}

	return dto
}

func ToContainerRegistry(dto ContainerRegistryDTO) *containerregistry.ContainerRegistry {
	cr := containerregistry.ContainerRegistry{
		Server:   dto.Server,
		Username: dto.Username,
		Password: dto.Password,
	}

	return &cr
}
