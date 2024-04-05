// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"github.com/daytonaio/daytona/pkg/apikey"
	. "github.com/daytonaio/daytona/pkg/db/dto"
	"gorm.io/gorm"
)

type ApiKeyStore struct {
	db *gorm.DB
}

func (a *ApiKeyStore) List() ([]*apikey.ApiKey, error) {
	apiKeyDTOs := []ApiKeyDTO{}
	tx := a.db.Find(&apiKeyDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	apiKeys := []*apikey.ApiKey{}
	for _, apiKeyDTO := range apiKeyDTOs {
		apiKey := ToApiKey(apiKeyDTO)
		apiKeys = append(apiKeys, &apiKey)
	}
	return apiKeys, nil
}

func (a *ApiKeyStore) Find(key string) (*apikey.ApiKey, error) {
	apiKeyDTO := ApiKeyDTO{}
	tx := a.db.Where("key_hash = ?", key).First(&apiKeyDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	apiKey := ToApiKey(apiKeyDTO)

	return &apiKey, nil
}

func (a *ApiKeyStore) FindByName(name string) (*apikey.ApiKey, error) {
	apiKeyDTO := ApiKeyDTO{}
	tx := a.db.Where("name = ?", name).First(&apiKeyDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	apiKey := ToApiKey(apiKeyDTO)

	return &apiKey, nil
}

func (a *ApiKeyStore) Save(apiKey *apikey.ApiKey) error {
	apiKeyDTO := ToApiKeyDTO(*apiKey)
	tx := a.db.Save(&apiKeyDTO)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (a *ApiKeyStore) Delete(apiKey *apikey.ApiKey) error {
	tx := a.db.Where("key_hash = ?", apiKey).Delete(&ApiKeyDTO{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
