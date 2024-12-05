// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type WorkspaceMetadataStore struct {
	Store
}

func NewWorkspaceMetadataStore(store Store) (stores.WorkspaceMetadataStore, error) {
	err := store.db.AutoMigrate(&models.WorkspaceMetadata{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceMetadataStore{store}, nil
}

func (s *WorkspaceMetadataStore) Find(ctx context.Context, filter *stores.WorkspaceMetadataFilter) (*models.WorkspaceMetadata, error) {
	tx := s.getTransaction(ctx)

	workspaceMetadata := &models.WorkspaceMetadata{}
	tx = processWorkspaceMetadataFilters(tx, filter).First(&workspaceMetadata)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrWorkspaceMetadataNotFound
		}
		return nil, tx.Error
	}

	return workspaceMetadata, nil
}

func (s *WorkspaceMetadataStore) Save(ctx context.Context, workspaceMetadata *models.WorkspaceMetadata) error {
	tx := s.getTransaction(ctx)

	tx = tx.Save(workspaceMetadata)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *WorkspaceMetadataStore) Delete(ctx context.Context, workspaceMetadata *models.WorkspaceMetadata) error {
	tx := s.getTransaction(ctx)

	tx = tx.Delete(workspaceMetadata)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrWorkspaceMetadataNotFound
	}

	return nil
}

func processWorkspaceMetadataFilters(tx *gorm.DB, filter *stores.WorkspaceMetadataFilter) *gorm.DB {
	if filter != nil {
		if filter.Id != nil {
			tx = tx.Where("id = ?", *filter.Id)
		}
		if filter.WorkspaceId != nil {
			tx = tx.Where("workspace_id = ?", *filter.WorkspaceId)
		}
	}
	return tx
}
