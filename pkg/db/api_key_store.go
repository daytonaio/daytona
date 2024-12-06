// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type ApiKeyStore struct {
	IStore
}

func NewApiKeyStore(store IStore) (stores.ApiKeyStore, error) {
	err := store.AutoMigrate(&models.ApiKey{})
	if err != nil {
		return nil, err
	}

	return &ApiKeyStore{store}, nil
}

func (a *ApiKeyStore) List(ctx context.Context) ([]*models.ApiKey, error) {
	tx := a.GetTransaction(ctx)

	apiKeys := []*models.ApiKey{}
	tx = tx.Find(&apiKeys)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return apiKeys, nil
}

func (a *ApiKeyStore) Find(ctx context.Context, key string) (*models.ApiKey, error) {
	tx := a.GetTransaction(ctx)

	apiKey := &models.ApiKey{}
	tx = tx.Where("key_hash = ?", key).First(apiKey)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrApiKeyNotFound
		}
		return nil, tx.Error
	}

	return apiKey, nil
}

func (a *ApiKeyStore) FindByName(ctx context.Context, name string) (*models.ApiKey, error) {
	tx := a.GetTransaction(ctx)

	apiKey := &models.ApiKey{}
	tx = tx.Where("name = ?", name).First(apiKey)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrApiKeyNotFound
		}
		return nil, tx.Error
	}

	return apiKey, nil
}

func (a *ApiKeyStore) Save(ctx context.Context, apiKey *models.ApiKey) error {
	tx := a.GetTransaction(ctx)

	tx = tx.Save(apiKey)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (a *ApiKeyStore) Delete(ctx context.Context, apiKey *models.ApiKey) error {
	tx := a.GetTransaction(ctx)

	tx = tx.Where("key_hash = ?", apiKey.KeyHash).Delete(&models.ApiKey{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrApiKeyNotFound
	}

	return nil
}
