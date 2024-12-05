//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/testing/common"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryWorkspaceTemplateStore struct {
	common.InMemoryStore
	workspaceTemplates map[string]*models.WorkspaceTemplate
}

func NewInMemoryWorkspaceTemplateStore() stores.WorkspaceTemplateStore {
	return &InMemoryWorkspaceTemplateStore{
		workspaceTemplates: make(map[string]*models.WorkspaceTemplate),
	}
}

func (s *InMemoryWorkspaceTemplateStore) List(ctx context.Context, filter *stores.WorkspaceTemplateFilter) ([]*models.WorkspaceTemplate, error) {
	return s.processFilters(filter)
}

func (s *InMemoryWorkspaceTemplateStore) Find(ctx context.Context, filter *stores.WorkspaceTemplateFilter) (*models.WorkspaceTemplate, error) {
	workspaceTemplates, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(workspaceTemplates) == 0 {
		return nil, stores.ErrWorkspaceTemplateNotFound
	}

	return workspaceTemplates[0], nil
}

func (s *InMemoryWorkspaceTemplateStore) Save(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate) error {
	s.workspaceTemplates[workspaceTemplate.Name] = workspaceTemplate
	return nil
}

func (s *InMemoryWorkspaceTemplateStore) Delete(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate) error {
	delete(s.workspaceTemplates, workspaceTemplate.Name)
	return nil
}

func (s *InMemoryWorkspaceTemplateStore) processFilters(filter *stores.WorkspaceTemplateFilter) ([]*models.WorkspaceTemplate, error) {
	var result []*models.WorkspaceTemplate
	filteredWorkspaceTemplates := make(map[string]*models.WorkspaceTemplate)
	for k, v := range s.workspaceTemplates {
		filteredWorkspaceTemplates[k] = v
	}

	if filter != nil {
		if filter.Name != nil {
			workspaceTemplate, ok := s.workspaceTemplates[*filter.Name]
			if ok {
				return []*models.WorkspaceTemplate{workspaceTemplate}, nil
			} else {
				return []*models.WorkspaceTemplate{}, fmt.Errorf("workspace template with name %s not found", *filter.Name)
			}
		}
		if filter.Url != nil {
			for _, workspaceTemplate := range filteredWorkspaceTemplates {
				if workspaceTemplate.RepositoryUrl != *filter.Url {
					delete(filteredWorkspaceTemplates, workspaceTemplate.Name)
				}
			}
		}
		if filter.Default != nil {
			for _, workspaceTemplate := range filteredWorkspaceTemplates {
				if workspaceTemplate.IsDefault != *filter.Default {
					delete(filteredWorkspaceTemplates, workspaceTemplate.Name)
				}
			}
		}
		if filter.GitProviderConfigId != nil {
			for _, workspaceTemplate := range filteredWorkspaceTemplates {
				if workspaceTemplate.GitProviderConfigId != nil && *workspaceTemplate.GitProviderConfigId != *filter.GitProviderConfigId {
					delete(filteredWorkspaceTemplates, workspaceTemplate.Name)
				}
			}
		}
	}

	for _, workspaceTemplate := range filteredWorkspaceTemplates {
		result = append(result, workspaceTemplate)
	}

	return result, nil
}
