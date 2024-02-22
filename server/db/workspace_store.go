// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/common/types"
)

func ListWorkspaces() ([]*types.Workspace, error) {
	db, err := getWorkspaceDB()
	if err != nil {
		return nil, err
	}

	workspaces := []*types.Workspace{}
	tx := db.Find(&workspaces)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return workspaces, nil
}

func FindWorkspace(workspaceId string) (*types.Workspace, error) {
	db, err := getWorkspaceDB()
	if err != nil {
		return nil, err
	}

	workspace := new(types.Workspace)
	tx := db.Where("id = ?", workspaceId).First(&workspace)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return workspace, nil
}

func SaveWorkspace(workspace *types.Workspace) error {
	db, err := getWorkspaceDB()
	if err != nil {
		return err
	}

	tx := db.Save(workspace)
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

	tx := db.Delete(workspace)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func getWorkspaceDB() (*gorm.DB, error) {
	db := getConnection()
	err := db.AutoMigrate(&types.Workspace{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
