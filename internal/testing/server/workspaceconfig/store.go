//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/target/workspace/config"
)

type InMemoryWorkspaceConfigStore struct {
	workspaceConfigs map[string]*config.WorkspaceConfig
}

func NewInMemoryWorkspaceConfigStore() config.Store {
	return &InMemoryWorkspaceConfigStore{
		workspaceConfigs: make(map[string]*config.WorkspaceConfig),
	}
}

func (s *InMemoryWorkspaceConfigStore) List(filter *config.WorkspaceConfigFilter) ([]*config.WorkspaceConfig, error) {
	return s.processFilters(filter)
}

func (s *InMemoryWorkspaceConfigStore) Find(filter *config.WorkspaceConfigFilter) (*config.WorkspaceConfig, error) {
	workspaceConfigs, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(workspaceConfigs) == 0 {
		return nil, config.ErrWorkspaceConfigNotFound
	}

	return workspaceConfigs[0], nil
}

func (s *InMemoryWorkspaceConfigStore) Save(workspaceConfig *config.WorkspaceConfig) error {
	s.workspaceConfigs[workspaceConfig.Name] = workspaceConfig
	return nil
}

func (s *InMemoryWorkspaceConfigStore) Delete(workspaceConfig *config.WorkspaceConfig) error {
	delete(s.workspaceConfigs, workspaceConfig.Name)
	return nil
}

func (s *InMemoryWorkspaceConfigStore) processFilters(filter *config.WorkspaceConfigFilter) ([]*config.WorkspaceConfig, error) {
	var result []*config.WorkspaceConfig
	filteredWorkspaceConfigs := make(map[string]*config.WorkspaceConfig)
	for k, v := range s.workspaceConfigs {
		filteredWorkspaceConfigs[k] = v
	}

	if filter != nil {
		if filter.Name != nil {
			workspaceConfig, ok := s.workspaceConfigs[*filter.Name]
			if ok {
				return []*config.WorkspaceConfig{workspaceConfig}, nil
			} else {
				return []*config.WorkspaceConfig{}, fmt.Errorf("workspace config with name %s not found", *filter.Name)
			}
		}
		if filter.Url != nil {
			for _, workspaceConfig := range filteredWorkspaceConfigs {
				if workspaceConfig.RepositoryUrl != *filter.Url {
					delete(filteredWorkspaceConfigs, workspaceConfig.Name)
				}
			}
		}
		if filter.Default != nil {
			for _, workspaceConfig := range filteredWorkspaceConfigs {
				if workspaceConfig.IsDefault != *filter.Default {
					delete(filteredWorkspaceConfigs, workspaceConfig.Name)
				}
			}
		}
		if filter.GitProviderConfigId != nil {
			for _, workspaceConfig := range filteredWorkspaceConfigs {
				if workspaceConfig.GitProviderConfigId != nil && *workspaceConfig.GitProviderConfigId != *filter.GitProviderConfigId {
					delete(filteredWorkspaceConfigs, workspaceConfig.Name)
				}
			}
		}
	}

	for _, workspaceConfig := range filteredWorkspaceConfigs {
		result = append(result, workspaceConfig)
	}

	return result, nil
}
