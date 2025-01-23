// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type WorkspaceStore struct {
	IStore
}

func NewWorkspaceStore(store IStore) (stores.WorkspaceStore, error) {
	err := store.AutoMigrate(&models.Workspace{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceStore{store}, nil
}

func (s *WorkspaceStore) List(ctx context.Context) ([]*models.Workspace, error) {
	tx := s.GetTransaction(ctx)

	workspaces := []*models.Workspace{}
	tx = preloadWorkspaceEntities(tx).Find(&workspaces)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return workspaces, nil
}

func (s *WorkspaceStore) Find(ctx context.Context, idOrName string) (*models.Workspace, error) {
	tx := s.GetTransaction(ctx)

	workspace := &models.Workspace{}
	tx = preloadWorkspaceEntities(tx).Where("id = ? OR name = ?", idOrName, idOrName).First(workspace)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrWorkspaceNotFound
		}
		return nil, tx.Error
	}

	return workspace, nil
}

func (s *WorkspaceStore) Save(ctx context.Context, workspace *models.Workspace) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Save(workspace)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *WorkspaceStore) Delete(ctx context.Context, workspace *models.Workspace) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Delete(workspace)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrWorkspaceNotFound
	}

	return nil
}

func preloadWorkspaceEntities(tx *gorm.DB) *gorm.DB {
	return tx.Preload(clause.Associations).Preload("Target.TargetConfig").Preload("LastJob")
}
