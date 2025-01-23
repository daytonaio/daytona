// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"context"

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

func (s *EnvironmentVariableService) List(ctx context.Context) ([]*models.EnvironmentVariable, error) {
	return s.environmentVariableStore.List(ctx)
}

func (s *EnvironmentVariableService) Map(ctx context.Context) (services.EnvironmentVariables, error) {
	envVars, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	envVarsMap := services.EnvironmentVariables{}
	for _, envVar := range envVars {
		envVarsMap[envVar.Key] = envVar.Value
	}

	return envVarsMap, nil
}

func (s *EnvironmentVariableService) Save(ctx context.Context, environmentVariable *models.EnvironmentVariable) error {
	return s.environmentVariableStore.Save(ctx, environmentVariable)
}

func (s *EnvironmentVariableService) Delete(ctx context.Context, key string) error {
	return s.environmentVariableStore.Delete(ctx, key)
}
