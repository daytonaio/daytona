// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strings"

	"github.com/daytonaio/daytona/pkg/models"
)

const (
	ContainerRegistryServerEnvVarSuffix   = "_CONTAINER_REGISTRY_SERVER"
	ContainerRegistryUsernameEnvVarSuffix = "_CONTAINER_REGISTRY_USERNAME"
	ContainerRegistryPasswordEnvVarSuffix = "_CONTAINER_REGISTRY_PASSWORD"
)

type ContainerRegistries map[string]*models.ContainerRegistry

func (c ContainerRegistries) FindContainerRegistryByImageName(image string) *models.ContainerRegistry {
	parts := strings.Split(image, "/")

	if len(parts) < 3 {
		return c["docker.io"]
	}

	return c[parts[0]]
}

func ExtractContainerRegistryFromEnvVars(envVars map[string]string) (map[string]string, ContainerRegistries) {
	resultEnvVars := make(map[string]string)
	containerRegistryEnvVars := make(map[string]*models.ContainerRegistry)

	for k, v := range envVars {
		if !isContainerRegistryEnvVar(k) {
			resultEnvVars[k] = v
		} else if strings.HasSuffix(k, ContainerRegistryServerEnvVarSuffix) {
			usernameKey := strings.ReplaceAll(k, ContainerRegistryServerEnvVarSuffix, ContainerRegistryUsernameEnvVarSuffix)
			passwordKey := strings.ReplaceAll(k, ContainerRegistryServerEnvVarSuffix, ContainerRegistryPasswordEnvVarSuffix)

			containerRegistryEnvVars[v] = &models.ContainerRegistry{
				Server:   v,
				Username: envVars[usernameKey],
				Password: envVars[passwordKey],
			}
		}
	}

	return resultEnvVars, containerRegistryEnvVars
}

func isContainerRegistryEnvVar(key string) bool {
	return strings.HasSuffix(key, ContainerRegistryServerEnvVarSuffix) || strings.HasSuffix(key, ContainerRegistryUsernameEnvVarSuffix) || strings.HasSuffix(key, ContainerRegistryPasswordEnvVarSuffix)
}
