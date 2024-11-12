// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"gorm.io/gorm"
)

type ApiKeyStore struct {
	db *gorm.DB
}

func NewApiKeyStore(db *gorm.DB) (*ApiKeyStore, error) {
	err := db.AutoMigrate(&models.ApiKey{})
	if err != nil {
		return nil, err
	}

	return &ApiKeyStore{db: db}, nil
}

func (a *ApiKeyStore) List() ([]*models.ApiKey, error) {
	apiKeys := []*models.ApiKey{}
	tx := a.db.Find(&apiKeys)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return apiKeys, nil
}

func (a *ApiKeyStore) Find(key string) (*models.ApiKey, error) {
	apiKey := &models.ApiKey{}
	tx := a.db.Where("key_hash = ?", key).First(apiKey)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, apikeys.ErrApiKeyNotFound
		}
		return nil, tx.Error
	}

	return apiKey, nil
}

func (a *ApiKeyStore) FindByName(name string) (*models.ApiKey, error) {
	apiKey := &models.ApiKey{}
	tx := a.db.Where("name = ?", name).First(apiKey)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, apikeys.ErrApiKeyNotFound
		}
		return nil, tx.Error
	}

	return apiKey, nil
}

func (a *ApiKeyStore) Save(apiKey *models.ApiKey) error {
	tx := a.db.Save(apiKey)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (a *ApiKeyStore) Delete(apiKey *models.ApiKey) error {
	tx := a.db.Where("key_hash = ?", apiKey.KeyHash).Delete(&models.ApiKey{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return apikeys.ErrApiKeyNotFound
	}

	return nil
}
