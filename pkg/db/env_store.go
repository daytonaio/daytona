// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"gorm.io/gorm"
)

type EnvironmentVariableStore struct {
	Store
}

func NewEnvironmentVariableStore(store Store) (stores.EnvironmentVariableStore, error) {
	err := store.db.AutoMigrate(&models.EnvironmentVariable{})
	if err != nil {
		return nil, err
	}

	return &EnvironmentVariableStore{store}, nil
}

func (store *EnvironmentVariableStore) List(ctx context.Context) ([]*models.EnvironmentVariable, error) {
	tx := store.getTransaction(ctx)

	environmentVariables := []*models.EnvironmentVariable{}
	tx = tx.Find(&environmentVariables)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, stores.ErrEnvironmentVariableNotFound
		}
		return nil, tx.Error
	}

	return environmentVariables, nil
}

func (store *EnvironmentVariableStore) Save(ctx context.Context, environmentVariable *models.EnvironmentVariable) error {
	tx := store.getTransaction(ctx)

	tx = tx.Save(environmentVariable)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (store *EnvironmentVariableStore) Delete(ctx context.Context, key string) error {
	tx := store.getTransaction(ctx)

	tx = tx.Where("key = ?", key).Delete(&models.EnvironmentVariable{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrEnvironmentVariableNotFound
	}

	return nil
}
