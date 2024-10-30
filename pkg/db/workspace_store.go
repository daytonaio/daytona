// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type WorkspaceStore struct {
	db *gorm.DB
}

func NewWorkspaceStore(db *gorm.DB) (*WorkspaceStore, error) {
	err := db.AutoMigrate(&WorkspaceDTO{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceStore{db: db}, nil
}

func (store *WorkspaceStore) List() ([]*workspace.WorkspaceViewDTO, error) {
	workspaceDTOs := []WorkspaceDTO{}
	tx := store.db.Preload(clause.Associations).Find(&workspaceDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	workspaceViewDTOs := []*workspace.WorkspaceViewDTO{}
	for _, workspaceDTO := range workspaceDTOs {
		workspaceViewDTOs = append(workspaceViewDTOs, ToWorkspaceViewDTO(workspaceDTO))
	}

	return workspaceViewDTOs, nil
}

func (w *WorkspaceStore) Find(idOrName string) (*workspace.WorkspaceViewDTO, error) {
	workspaceDTO := WorkspaceDTO{}
	tx := w.db.Preload(clause.Associations).Where("id = ? OR name = ?", idOrName, idOrName).First(&workspaceDTO)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, workspace.ErrWorkspaceNotFound
		}
		return nil, tx.Error
	}

	return ToWorkspaceViewDTO(workspaceDTO), nil
}

func (w *WorkspaceStore) Save(workspace *workspace.Workspace) error {
	tx := w.db.Save(ToWorkspaceDTO(workspace))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (w *WorkspaceStore) Delete(t *workspace.Workspace) error {
	tx := w.db.Delete(ToWorkspaceDTO(t))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return workspace.ErrWorkspaceNotFound
	}

	return nil
}
