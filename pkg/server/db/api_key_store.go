// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	. "github.com/daytonaio/daytona/pkg/server/db/dto"
	"github.com/daytonaio/daytona/pkg/types"
	"gorm.io/gorm"
)

func ListApiKeys() ([]*types.ApiKey, error) {
	db, err := getApiKeyDB()
	if err != nil {
		return nil, err
	}

	apiKeyDTOs := []ApiKeyDTO{}
	tx := db.Find(&apiKeyDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	apiKeys := []*types.ApiKey{}
	for _, apiKeyDTO := range apiKeyDTOs {
		apiKey := ToApiKey(apiKeyDTO)
		apiKeys = append(apiKeys, &apiKey)
	}
	return apiKeys, nil
}

func FindApiKey(key string) (*types.ApiKey, error) {
	db, err := getApiKeyDB()
	if err != nil {
		return nil, err
	}

	apiKeyDTO := ApiKeyDTO{}
	tx := db.Where("key_hash = ?", key).First(&apiKeyDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	apiKey := ToApiKey(apiKeyDTO)

	return &apiKey, nil
}

func FindApiKeyByName(name string) (*types.ApiKey, error) {
	db, err := getApiKeyDB()
	if err != nil {
		return nil, err
	}

	apiKeyDTO := ApiKeyDTO{}
	tx := db.Where("name = ?", name).First(&apiKeyDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	apiKey := ToApiKey(apiKeyDTO)

	return &apiKey, nil
}

func SaveApiKey(apiKey *types.ApiKey) error {
	db, err := getApiKeyDB()
	if err != nil {
		return err
	}

	apiKeyDTO := ToApiKeyDTO(*apiKey)
	tx := db.Save(&apiKeyDTO)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func DeleteApiKey(key string) error {
	db, err := getApiKeyDB()
	if err != nil {
		return err
	}

	tx := db.Where("key_hash = ?", key).Delete(&ApiKeyDTO{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func getApiKeyDB() (*gorm.DB, error) {
	db := getConnection()
	err := db.AutoMigrate(&ApiKeyDTO{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
