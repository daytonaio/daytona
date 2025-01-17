//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"context"

	"github.com/daytonaio/daytona/internal/testing/common"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryEnvironmentVariableStore struct {
	common.InMemoryStore
	envVars map[string]*models.EnvironmentVariable
}

func NewInMemoryEnvironmentVariableStore() stores.EnvironmentVariableStore {
	return &InMemoryEnvironmentVariableStore{
		envVars: make(map[string]*models.EnvironmentVariable),
	}
}

func (s *InMemoryEnvironmentVariableStore) List(ctx context.Context) ([]*models.EnvironmentVariable, error) {
	envVars := []*models.EnvironmentVariable{}
	for _, envVar := range s.envVars {
		envVars = append(envVars, envVar)
	}

	return envVars, nil
}

func (s *InMemoryEnvironmentVariableStore) Save(ctx context.Context, environmentVariable *models.EnvironmentVariable) error {
	s.envVars[environmentVariable.Key] = environmentVariable
	return nil
}

func (s *InMemoryEnvironmentVariableStore) Delete(ctx context.Context, key string) error {
	_, ok := s.envVars[key]
	if !ok {
		return stores.ErrEnvironmentVariableNotFound
	}
	delete(s.envVars, key)
	return nil
}
