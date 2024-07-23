//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

type InMemoryProjectConfigStore struct {
	projectConfigs map[string]*config.ProjectConfig
}

func NewInMemoryProjectConfigStore() config.Store {
	return &InMemoryProjectConfigStore{
		projectConfigs: make(map[string]*config.ProjectConfig),
	}
}

func (s *InMemoryProjectConfigStore) List() ([]*config.ProjectConfig, error) {
	projectConfigs := []*config.ProjectConfig{}
	for _, t := range s.projectConfigs {
		projectConfigs = append(projectConfigs, t)
	}

	return projectConfigs, nil
}

func (s *InMemoryProjectConfigStore) Find(projectConfigName string) (*config.ProjectConfig, error) {
	projectConfig, ok := s.projectConfigs[projectConfigName]
	if !ok {
		return nil, config.ErrProjectConfigNotFound
	}

	return projectConfig, nil
}

func (s *InMemoryProjectConfigStore) Save(projectConfig *config.ProjectConfig) error {
	s.projectConfigs[projectConfig.Name] = projectConfig
	return nil
}

func (s *InMemoryProjectConfigStore) Delete(projectConfig *config.ProjectConfig) error {
	delete(s.projectConfigs, projectConfig.Name)
	return nil
}
