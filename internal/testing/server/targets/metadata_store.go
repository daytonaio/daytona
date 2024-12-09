//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/internal/testing/common"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type InMemoryTargetMetadataStore struct {
	common.InMemoryStore
	targetMetadataEntries map[string]*models.TargetMetadata
}

func NewInMemoryTargetMetadataStore() stores.TargetMetadataStore {
	return &InMemoryTargetMetadataStore{
		targetMetadataEntries: make(map[string]*models.TargetMetadata),
	}
}

func (s *InMemoryTargetMetadataStore) Find(ctx context.Context, targetId string) (*models.TargetMetadata, error) {
	metadata, ok := s.targetMetadataEntries[targetId]
	if !ok {
		return nil, stores.ErrTargetMetadataNotFound
	}

	return metadata, nil
}

func (s *InMemoryTargetMetadataStore) Save(ctx context.Context, targetMetadata *models.TargetMetadata) error {
	s.targetMetadataEntries[targetMetadata.TargetId] = targetMetadata
	return nil
}

func (s *InMemoryTargetMetadataStore) Delete(ctx context.Context, targetMetadata *models.TargetMetadata) error {
	delete(s.targetMetadataEntries, targetMetadata.TargetId)
	return nil
}
