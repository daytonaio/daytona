// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
)

type GitProviderConfigStore struct {
	db *gorm.DB
}

func NewGitProviderConfigStore(db *gorm.DB) (*GitProviderConfigStore, error) {
	err := db.AutoMigrate(&models.GitProviderConfig{})
	if err != nil {
		return nil, err
	}

	return &GitProviderConfigStore{db: db}, nil
}

func (p *GitProviderConfigStore) List() ([]*models.GitProviderConfig, error) {
	gitProviders := []*models.GitProviderConfig{}
	tx := p.db.Find(&gitProviders)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return gitProviders, nil
}

func (p *GitProviderConfigStore) Find(id string) (*models.GitProviderConfig, error) {
	gitProvider := &models.GitProviderConfig{}
	tx := p.db.Where("id = ?", id).First(gitProvider)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, gitproviders.ErrGitProviderConfigNotFound
		}
		return nil, tx.Error
	}

	return gitProvider, nil
}

func (p *GitProviderConfigStore) Save(gitProvider *models.GitProviderConfig) error {
	tx := p.db.Save(gitProvider)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (p *GitProviderConfigStore) Delete(gitProvider *models.GitProviderConfig) error {
	tx := p.db.Delete(gitProvider)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gitproviders.ErrGitProviderConfigNotFound
	}

	return nil
}
