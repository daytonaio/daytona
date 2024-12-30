// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

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

func (w *WorkspaceStore) List() ([]*workspace.Workspace, error) {
	workspaceDTOs := []WorkspaceDTO{}
	tx := w.db.Find(&workspaceDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	workspaces := []*workspace.Workspace{}
	for _, workspaceDTO := range workspaceDTOs {
		workspaces = append(workspaces, ToWorkspace(workspaceDTO))
	}

	return workspaces, nil
}

func (w *WorkspaceStore) Find(idOrName string) (*workspace.Workspace, error) {
	workspaceDTO := WorkspaceDTO{}
	tx := w.db.Where("id = ? OR name = ?", idOrName, idOrName).First(&workspaceDTO)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, workspace.ErrWorkspaceNotFound
		}
		return nil, tx.Error
	}

	return ToWorkspace(workspaceDTO), nil
}

func (w *WorkspaceStore) Save(workspace *workspace.Workspace) error {
	tx := w.db.Save(ToWorkspaceDTO(workspace))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (w *WorkspaceStore) Delete(ws *workspace.Workspace) error {
	tx := w.db.Delete(ToWorkspaceDTO(ws))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return workspace.ErrWorkspaceNotFound
	}

	return nil
}
