// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
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

func (store *WorkspaceStore) List() ([]*models.Workspace, error) {
	workspaces := []*models.Workspace{}
	tx := store.db.Preload(clause.Associations).Find(&workspaces)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return workspaces, nil
}

func (w *WorkspaceStore) Find(idOrName string) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	tx := w.db.Preload(clause.Associations).Where("id = ? OR name = ?", idOrName, idOrName).First(workspace)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, workspaces.ErrWorkspaceNotFound
		}
		return nil, tx.Error
	}

	return workspace, nil
}

func (w *WorkspaceStore) Save(workspace *models.Workspace) error {
	tx := w.db.Save(workspace)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (w *WorkspaceStore) Delete(workspace *models.Workspace) error {
	tx := w.db.Delete(workspace)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return workspaces.ErrWorkspaceNotFound
	}

	return nil
}
