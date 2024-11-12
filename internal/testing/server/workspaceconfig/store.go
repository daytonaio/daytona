//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfigs"
)

type InMemoryWorkspaceConfigStore struct {
	workspaceConfigs map[string]*models.WorkspaceConfig
}

func NewInMemoryWorkspaceConfigStore() workspaceconfigs.WorkspaceConfigStore {
	return &InMemoryWorkspaceConfigStore{
		workspaceConfigs: make(map[string]*models.WorkspaceConfig),
	}
}

func (s *InMemoryWorkspaceConfigStore) List(filter *workspaceconfigs.WorkspaceConfigFilter) ([]*models.WorkspaceConfig, error) {
	return s.processFilters(filter)
}

func (s *InMemoryWorkspaceConfigStore) Find(filter *workspaceconfigs.WorkspaceConfigFilter) (*models.WorkspaceConfig, error) {
	workspaceConfigs, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(workspaceConfigs) == 0 {
		return nil, workspaceconfigs.ErrWorkspaceConfigNotFound
	}

	return workspaceConfigs[0], nil
}

func (s *InMemoryWorkspaceConfigStore) Save(workspaceConfig *models.WorkspaceConfig) error {
	s.workspaceConfigs[workspaceConfig.Name] = workspaceConfig
	return nil
}

func (s *InMemoryWorkspaceConfigStore) Delete(workspaceConfig *models.WorkspaceConfig) error {
	delete(s.workspaceConfigs, workspaceConfig.Name)
	return nil
}

func (s *InMemoryWorkspaceConfigStore) processFilters(filter *workspaceconfigs.WorkspaceConfigFilter) ([]*models.WorkspaceConfig, error) {
	var result []*models.WorkspaceConfig
	filteredWorkspaceConfigs := make(map[string]*models.WorkspaceConfig)
	for k, v := range s.workspaceConfigs {
		filteredWorkspaceConfigs[k] = v
	}

	if filter != nil {
		if filter.Name != nil {
			workspaceConfig, ok := s.workspaceConfigs[*filter.Name]
			if ok {
				return []*models.WorkspaceConfig{workspaceConfig}, nil
			} else {
				return []*models.WorkspaceConfig{}, fmt.Errorf("workspace config with name %s not found", *filter.Name)
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
