// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"strings"

	"github.com/daytonaio/daytona/pkg/models"
)

type IEnvironmentVariableService interface {
	List() ([]*models.EnvironmentVariable, error)
	Map() (EnvironmentVariables, error)
	Save(environmentVariable *models.EnvironmentVariable) error
	Delete(key string) error
}

type EnvironmentVariables map[string]string

func (e EnvironmentVariables) FindContainerRegistry(server string) *models.ContainerRegistry {
	for key, value := range e {
		if strings.HasSuffix(key, "CONTAINER_REGISTRY_SERVER") && value == server {
			usernameKey := strings.ReplaceAll(key, "SERVER", "USERNAME")
			passwordKey := strings.ReplaceAll(key, "SERVER", "PASSWORD")

			return &models.ContainerRegistry{
				Server:   server,
				Username: e[usernameKey],
				Password: e[passwordKey],
			}
		}
	}

	return nil
}

func (e EnvironmentVariables) FindContainerRegistryByImageName(image string) *models.ContainerRegistry {
	parts := strings.Split(image, "/")

	if len(parts) < 3 {
		return e.FindContainerRegistry("docker.io")
	}

	return e.FindContainerRegistry(parts[0])
}
