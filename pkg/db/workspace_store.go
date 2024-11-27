// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type WorkspaceStore struct {
	db *gorm.DB
}

func NewWorkspaceStore(db *gorm.DB) (*WorkspaceStore, error) {
	err := db.AutoMigrate(&models.Workspace{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceStore{db: db}, nil
}

func (s *WorkspaceStore) List() ([]*models.Workspace, error) {
	workspaces := []*models.Workspace{}
	// Order workspace jobs by created_at
	tx := preloadWorkspaceEntities(s.db).Find(&workspaces)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return workspaces, nil
}

func (s *WorkspaceStore) Find(idOrName string) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	tx := preloadWorkspaceEntities(s.db).Where("id = ? OR name = ?", idOrName, idOrName).First(workspace)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrWorkspaceNotFound
		}
		return nil, tx.Error
	}

	return workspace, nil
}

func (s *WorkspaceStore) Save(workspace *models.Workspace) error {
	tx := s.db.Save(workspace)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *WorkspaceStore) Delete(workspace *models.Workspace) error {
	tx := s.db.Delete(workspace)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrWorkspaceNotFound
	}

	return nil
}

func preloadWorkspaceEntities(tx *gorm.DB) *gorm.DB {
	return tx.Preload(clause.Associations).Preload("LastJob", preloadLastJob)
}
