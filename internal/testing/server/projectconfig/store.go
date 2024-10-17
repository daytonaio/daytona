//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"fmt"

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

func (s *InMemoryProjectConfigStore) List(filter *config.ProjectConfigFilter) ([]*config.ProjectConfig, error) {
	return s.processFilters(filter)
}

func (s *InMemoryProjectConfigStore) Find(filter *config.ProjectConfigFilter) (*config.ProjectConfig, error) {
	projectConfigs, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(projectConfigs) == 0 {
		return nil, config.ErrProjectConfigNotFound
	}

	return projectConfigs[0], nil
}

func (s *InMemoryProjectConfigStore) Save(projectConfig *config.ProjectConfig) error {
	s.projectConfigs[projectConfig.Name] = projectConfig
	return nil
}

func (s *InMemoryProjectConfigStore) Delete(projectConfig *config.ProjectConfig) error {
	delete(s.projectConfigs, projectConfig.Name)
	return nil
}

func (s *InMemoryProjectConfigStore) processFilters(filter *config.ProjectConfigFilter) ([]*config.ProjectConfig, error) {
	var result []*config.ProjectConfig
	filteredProjectConfigs := make(map[string]*config.ProjectConfig)
	for k, v := range s.projectConfigs {
		filteredProjectConfigs[k] = v
	}

	if filter != nil {
		if filter.Name != nil {
			projectConfig, ok := s.projectConfigs[*filter.Name]
			if ok {
				return []*config.ProjectConfig{projectConfig}, nil
			} else {
				return []*config.ProjectConfig{}, fmt.Errorf("project config with name %s not found", *filter.Name)
			}
		}
		if filter.Url != nil {
			for _, projectConfig := range filteredProjectConfigs {
				if projectConfig.RepositoryUrl != *filter.Url {
					delete(filteredProjectConfigs, projectConfig.Name)
				}
			}
		}
		if filter.Default != nil {
			for _, projectConfig := range filteredProjectConfigs {
				if projectConfig.IsDefault != *filter.Default {
					delete(filteredProjectConfigs, projectConfig.Name)
				}
			}
		}
		if filter.GitProviderConfigId != nil {
			for _, projectConfig := range filteredProjectConfigs {
				if projectConfig.GitProviderConfigId != nil && *projectConfig.GitProviderConfigId != *filter.GitProviderConfigId {
					delete(filteredProjectConfigs, projectConfig.Name)
				}
			}
		}
	}

	for _, projectConfig := range filteredProjectConfigs {
		result = append(result, projectConfig)
	}

	return result, nil
}
