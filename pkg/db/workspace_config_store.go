// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfigs"
)

type WorkspaceConfigStore struct {
	db *gorm.DB
}

func NewWorkspaceConfigStore(db *gorm.DB) (*WorkspaceConfigStore, error) {
	err := db.AutoMigrate(&models.WorkspaceConfig{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceConfigStore{db: db}, nil
}

func (s *WorkspaceConfigStore) List(filter *workspaceconfigs.WorkspaceConfigFilter) ([]*models.WorkspaceConfig, error) {
	workspaceConfigs := []*models.WorkspaceConfig{}
	tx := processWorkspaceConfigFilters(s.db, filter).Find(&workspaceConfigs)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return workspaceConfigs, nil
}

func (s *WorkspaceConfigStore) Find(filter *workspaceconfigs.WorkspaceConfigFilter) (*models.WorkspaceConfig, error) {
	workspaceConfig := &models.WorkspaceConfig{}
	tx := processWorkspaceConfigFilters(s.db, filter).First(workspaceConfig)

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, workspaceconfigs.ErrWorkspaceConfigNotFound
		}
		return nil, tx.Error
	}

	return workspaceConfig, nil
}

func (s *WorkspaceConfigStore) Save(workspaceConfig *models.WorkspaceConfig) error {
	tx := s.db.Save(workspaceConfig)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *WorkspaceConfigStore) Delete(workspaceConfig *models.WorkspaceConfig) error {
	tx := s.db.Delete(workspaceConfig)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return workspaceconfigs.ErrWorkspaceConfigNotFound
	}

	return nil
}

func processWorkspaceConfigFilters(tx *gorm.DB, filter *workspaceconfigs.WorkspaceConfigFilter) *gorm.DB {
	if filter != nil {
		if filter.Name != nil {
			tx = tx.Where("name = ?", *filter.Name)
		}
		if filter.Url != nil {
			tx = tx.Where("repository_url = ?", *filter.Url)
		}
		if filter.Default != nil {
			tx = tx.Where("is_default = ?", *filter.Default)
		}
		if filter.GitProviderConfigId != nil {
			tx = tx.Where("git_provider_config_id = ?", *filter.GitProviderConfigId)
		}
	}

	return tx
}
