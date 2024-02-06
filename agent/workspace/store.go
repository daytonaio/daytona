// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"github.com/daytonaio/daytona/agent/db"

	"gorm.io/gorm"
)

func ListFromDB() ([]Workspace, error) {
	db, err := getConnection()
	if err != nil {
		return nil, err
	}

	workspaces := []Workspace{}
	tx := db.Find(&workspaces)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for i, _ := range workspaces {
		for j, _ := range workspaces[i].Projects {
			workspaces[i].Projects[j].Workspace = &workspaces[i]
		}
	}

	return workspaces, nil
}

func LoadFromDB(workspaceName string) (*Workspace, error) {
	db, err := getConnection()
	if err != nil {
		return nil, err
	}

	workspace := Workspace{}
	tx := db.Where("name = ?", workspaceName).First(&workspace)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for i, _ := range workspace.Projects {
		workspace.Projects[i].Workspace = &workspace
	}

	return &workspace, nil
}

func SaveToDB(workspace *Workspace) error {
	db, err := getConnection()
	if err != nil {
		return err
	}

	tx := db.Save(workspace)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func DeleteFromDB(workspace *Workspace) error {
	db, err := getConnection()
	if err != nil {
		return err
	}

	tx := db.Delete(workspace)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func getConnection() (*gorm.DB, error) {
	db := db.GetConnection()
	err := db.AutoMigrate(&Workspace{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
