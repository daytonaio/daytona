// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"gorm.io/gorm"
)

type EnvironmentVariableStore struct {
	db *gorm.DB
}

func NewEnvironmentVariableStore(db *gorm.DB) (*EnvironmentVariableStore, error) {
	err := db.AutoMigrate(&models.EnvironmentVariable{})
	if err != nil {
		return nil, err
	}

	return &EnvironmentVariableStore{db: db}, nil
}

func (store *EnvironmentVariableStore) List() ([]*models.EnvironmentVariable, error) {
	environmentVariables := []*models.EnvironmentVariable{}
	tx := store.db.Find(&environmentVariables)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, stores.ErrEnvironmentVariableNotFound
		}
		return nil, tx.Error
	}

	return environmentVariables, nil
}

func (store *EnvironmentVariableStore) Save(environmentVariable *models.EnvironmentVariable) error {
	tx := store.db.Save(environmentVariable)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (store *EnvironmentVariableStore) Delete(key string) error {
	tx := store.db.Where("key = ?", key).Delete(&models.EnvironmentVariable{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrEnvironmentVariableNotFound
	}

	return nil
}
