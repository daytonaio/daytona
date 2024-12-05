//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"

	"github.com/daytonaio/daytona/internal/testing/common"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryWorkspaceMetadataStore struct {
	common.InMemoryStore
	workspaceMetadataEntries map[string]*models.WorkspaceMetadata
}

func NewInMemoryWorkspaceMetadataStore() stores.WorkspaceMetadataStore {
	return &InMemoryWorkspaceMetadataStore{
		workspaceMetadataEntries: make(map[string]*models.WorkspaceMetadata),
	}
}

func (s *InMemoryWorkspaceMetadataStore) Find(ctx context.Context, filter *stores.WorkspaceMetadataFilter) (*models.WorkspaceMetadata, error) {
	metadatas, err := s.processFilters(filter)
	if err != nil {
		return nil, err
	}
	if len(metadatas) == 0 {
		return nil, stores.ErrWorkspaceMetadataNotFound
	}

	return metadatas[0], nil
}

func (s *InMemoryWorkspaceMetadataStore) Save(ctx context.Context, workspaceMetadata *models.WorkspaceMetadata) error {
	s.workspaceMetadataEntries[workspaceMetadata.WorkspaceId] = workspaceMetadata
	return nil
}

func (s *InMemoryWorkspaceMetadataStore) Delete(ctx context.Context, workspaceMetadata *models.WorkspaceMetadata) error {
	delete(s.workspaceMetadataEntries, workspaceMetadata.WorkspaceId)
	return nil
}

func (s *InMemoryWorkspaceMetadataStore) processFilters(filter *stores.WorkspaceMetadataFilter) ([]*models.WorkspaceMetadata, error) {
	var result []*models.WorkspaceMetadata
	filteredWorkspaceMetadata := make(map[string]*models.WorkspaceMetadata)
	for k, v := range s.workspaceMetadataEntries {
		filteredWorkspaceMetadata[k] = v
	}

	if filter != nil {
		if filter.Id != nil {
			m, ok := s.workspaceMetadataEntries[*filter.Id]
			if ok {
				return []*models.WorkspaceMetadata{m}, nil
			} else {
				return []*models.WorkspaceMetadata{}, nil
			}
		}
		if filter.WorkspaceId != nil {
			for _, m := range filteredWorkspaceMetadata {
				if m.WorkspaceId != *filter.WorkspaceId {
					delete(filteredWorkspaceMetadata, m.WorkspaceId)
				}
			}
		}
	}

	for _, m := range filteredWorkspaceMetadata {
		result = append(result, m)
	}

	return result, nil
}
