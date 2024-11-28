// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

type EnvironmentVariableServiceConfig struct {
	EnvironmentVariableStore stores.EnvironmentVariableStore
}

func NewEnvironmentVariableService(config EnvironmentVariableServiceConfig) services.IEnvironmentVariableService {
	return &EnvironmentVariableService{
		environmentVariableStore: config.EnvironmentVariableStore,
	}
}

type EnvironmentVariableService struct {
	environmentVariableStore stores.EnvironmentVariableStore
}

func (s *EnvironmentVariableService) List() ([]*models.EnvironmentVariable, error) {
	return s.environmentVariableStore.List()
}

func (s *EnvironmentVariableService) Save(environmentVariable *models.EnvironmentVariable) error {
	return s.environmentVariableStore.Save(environmentVariable)
}

func (s *EnvironmentVariableService) Delete(key string) error {
	return s.environmentVariableStore.Delete(key)
}
