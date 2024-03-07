// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	. "github.com/daytonaio/daytona/pkg/server/db/dto"
	"github.com/daytonaio/daytona/pkg/types"
)

func ListWorkspaces() ([]*types.Workspace, error) {
	db, err := getWorkspaceDB()
	if err != nil {
		return nil, err
	}

	workspaceDTOs := []WorkspaceDTO{}
	tx := db.Find(&workspaceDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	workspaces := []*types.Workspace{}
	for _, workspace := range workspaceDTOs {
		workspaces = append(workspaces, ToWorkspace(workspace))
	}

	return workspaces, nil
}

func FindWorkspaceById(workspaceId string) (*types.Workspace, error) {
	db, err := getWorkspaceDB()
	if err != nil {
		return nil, err
	}

	workspaceDTO := WorkspaceDTO{}
	tx := db.Where("id = ?", workspaceId).First(&workspaceDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return ToWorkspace(workspaceDTO), nil
}

func FindWorkspaceByName(workspaceName string) (*types.Workspace, error) {
	db, err := getWorkspaceDB()
	if err != nil {
		return nil, err
	}

	workspaceDTO := WorkspaceDTO{}
	tx := db.Where("name = ?", workspaceName).First(&workspaceDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return ToWorkspace(workspaceDTO), nil
}

func FindWorkspaceByIdOrName(workspaceIdOrName string) (*types.Workspace, error) {
	db, err := getWorkspaceDB()
	if err != nil {
		return nil, err
	}

	workspaceDTO := WorkspaceDTO{}
	tx := db.Where("id = ? OR name = ?", workspaceIdOrName, workspaceIdOrName).First(&workspaceDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return ToWorkspace(workspaceDTO), nil
}

func SaveWorkspace(workspace *types.Workspace) error {
	db, err := getWorkspaceDB()
	if err != nil {
		return err
	}

	tx := db.Save(ToWorkspaceDTO(workspace))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func DeleteWorkspace(workspace *types.Workspace) error {
	db, err := getWorkspaceDB()
	if err != nil {
		return err
	}

	tx := db.Delete(ToWorkspaceDTO(workspace))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func getWorkspaceDB() (*gorm.DB, error) {
	db := getConnection()
	err := db.AutoMigrate(&WorkspaceDTO{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
