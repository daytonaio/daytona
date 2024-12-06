// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type GitProviderConfigStore struct {
	IStore
}

func NewGitProviderConfigStore(store IStore) (stores.GitProviderConfigStore, error) {
	err := store.AutoMigrate(&models.GitProviderConfig{})
	if err != nil {
		return nil, err
	}

	return &GitProviderConfigStore{store}, nil
}

func (p *GitProviderConfigStore) List(ctx context.Context) ([]*models.GitProviderConfig, error) {
	tx := p.GetTransaction(ctx)

	gitProviders := []*models.GitProviderConfig{}
	tx = tx.Find(&gitProviders)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return gitProviders, nil
}

func (p *GitProviderConfigStore) Find(ctx context.Context, id string) (*models.GitProviderConfig, error) {
	tx := p.GetTransaction(ctx)

	gitProvider := &models.GitProviderConfig{}
	tx = tx.Where("id = ?", id).First(gitProvider)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrGitProviderConfigNotFound
		}
		return nil, tx.Error
	}

	return gitProvider, nil
}

func (p *GitProviderConfigStore) Save(ctx context.Context, gitProvider *models.GitProviderConfig) error {
	tx := p.GetTransaction(ctx)

	tx = tx.Save(gitProvider)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (p *GitProviderConfigStore) Delete(ctx context.Context, gitProvider *models.GitProviderConfig) error {
	tx := p.GetTransaction(ctx)

	tx = tx.Delete(gitProvider)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrGitProviderConfigNotFound
	}

	return nil
}
