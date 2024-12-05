// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type GitProviderConfigStore struct {
	Store
}

func NewGitProviderConfigStore(store Store) (stores.GitProviderConfigStore, error) {
	err := store.db.AutoMigrate(&models.GitProviderConfig{})
	if err != nil {
		return nil, err
	}

	return &GitProviderConfigStore{store}, nil
}

func (p *GitProviderConfigStore) List(ctx context.Context) ([]*models.GitProviderConfig, error) {
	tx := p.getTransaction(ctx)

	gitProviders := []*models.GitProviderConfig{}
	tx = tx.Find(&gitProviders)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return gitProviders, nil
}

func (p *GitProviderConfigStore) Find(ctx context.Context, id string) (*models.GitProviderConfig, error) {
	tx := p.getTransaction(ctx)

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
	tx := p.getTransaction(ctx)

	tx = tx.Save(gitProvider)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (p *GitProviderConfigStore) Delete(ctx context.Context, gitProvider *models.GitProviderConfig) error {
	tx := p.getTransaction(ctx)

	tx = tx.Delete(gitProvider)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrGitProviderConfigNotFound
	}

	return nil
}
