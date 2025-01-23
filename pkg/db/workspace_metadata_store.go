// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type WorkspaceMetadataStore struct {
	IStore
}

func NewWorkspaceMetadataStore(store IStore) (stores.WorkspaceMetadataStore, error) {
	err := store.AutoMigrate(&models.WorkspaceMetadata{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceMetadataStore{store}, nil
}

func (s *WorkspaceMetadataStore) Find(ctx context.Context, workspaceId string) (*models.WorkspaceMetadata, error) {
	tx := s.GetTransaction(ctx)

	workspaceMetadata := &models.WorkspaceMetadata{}
	tx = tx.Where("workspace_id = ?", workspaceId).First(&workspaceMetadata)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrWorkspaceMetadataNotFound
		}
		return nil, tx.Error
	}

	return workspaceMetadata, nil
}

func (s *WorkspaceMetadataStore) Save(ctx context.Context, workspaceMetadata *models.WorkspaceMetadata) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Save(workspaceMetadata)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *WorkspaceMetadataStore) Delete(ctx context.Context, workspaceMetadata *models.WorkspaceMetadata) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Delete(workspaceMetadata)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrWorkspaceMetadataNotFound
	}

	return nil
}
