// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type WorkspaceMetadataStore struct {
	db *gorm.DB
}

func NewWorkspaceMetadataStore(db *gorm.DB) (*WorkspaceMetadataStore, error) {
	err := db.AutoMigrate(&models.WorkspaceMetadata{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceMetadataStore{db: db}, nil
}

func (s *WorkspaceMetadataStore) Find(filter *stores.WorkspaceMetadataFilter) (*models.WorkspaceMetadata, error) {
	workspaceMetadata := &models.WorkspaceMetadata{}
	tx := processWorkspaceMetadataFilters(s.db, filter).First(&workspaceMetadata)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrWorkspaceMetadataNotFound
		}
		return nil, tx.Error
	}

	return workspaceMetadata, nil
}

func (s *WorkspaceMetadataStore) Save(workspaceMetadata *models.WorkspaceMetadata) error {
	tx := s.db.Save(workspaceMetadata)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *WorkspaceMetadataStore) Delete(workspaceMetadata *models.WorkspaceMetadata) error {
	tx := s.db.Delete(workspaceMetadata)
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
