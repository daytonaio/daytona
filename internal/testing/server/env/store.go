//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryEnvironmentVariableStore struct {
	envVars map[string]*models.EnvironmentVariable
}

func NewInMemoryEnvironmentVariableStore() stores.EnvironmentVariableStore {
	return &InMemoryEnvironmentVariableStore{
		envVars: make(map[string]*models.EnvironmentVariable),
	}
}

func (s *InMemoryEnvironmentVariableStore) List() ([]*models.EnvironmentVariable, error) {
	envVars := []*models.EnvironmentVariable{}
	for _, envVar := range s.envVars {
		envVars = append(envVars, envVar)
	}

	return envVars, nil
}

func (s *InMemoryEnvironmentVariableStore) Save(environmentVariable *models.EnvironmentVariable) error {
	s.envVars[environmentVariable.Key] = environmentVariable
	return nil
}

func (s *InMemoryEnvironmentVariableStore) Delete(key string) error {
	_, ok := s.envVars[key]
	if !ok {
		return stores.ErrEnvironmentVariableNotFound
	}
	delete(s.envVars, key)
	return nil
}
