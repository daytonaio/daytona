// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/workspace/config"
)

type WorkspaceConfigStore struct {
	db *gorm.DB
}

func NewWorkspaceConfigStore(db *gorm.DB) (*WorkspaceConfigStore, error) {
	err := db.AutoMigrate(&WorkspaceConfigDTO{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceConfigStore{db: db}, nil
}

func (s *WorkspaceConfigStore) List(filter *config.WorkspaceConfigFilter) ([]*config.WorkspaceConfig, error) {
	workspaceConfigDTOs := []WorkspaceConfigDTO{}
	tx := processWorkspaceConfigFilters(s.db, filter).Find(&workspaceConfigDTOs)

	if tx.Error != nil {
		return nil, tx.Error
	}

	workspaceConfigs := []*config.WorkspaceConfig{}
	for _, workspaceConfigDTO := range workspaceConfigDTOs {
		workspaceConfigs = append(workspaceConfigs, ToWorkspaceConfig(workspaceConfigDTO))
	}

	return workspaceConfigs, nil
}

func (s *WorkspaceConfigStore) Find(filter *config.WorkspaceConfigFilter) (*config.WorkspaceConfig, error) {
	workspaceConfigDTO := WorkspaceConfigDTO{}
	tx := processWorkspaceConfigFilters(s.db, filter).First(&workspaceConfigDTO)

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, config.ErrWorkspaceConfigNotFound
		}
		return nil, tx.Error
	}

	return ToWorkspaceConfig(workspaceConfigDTO), nil
}

func (s *WorkspaceConfigStore) Save(workspaceConfig *config.WorkspaceConfig) error {
	tx := s.db.Save(ToWorkspaceConfigDTO(workspaceConfig))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *WorkspaceConfigStore) Delete(workspaceConfig *config.WorkspaceConfig) error {
	tx := s.db.Delete(ToWorkspaceConfigDTO(workspaceConfig))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return config.ErrWorkspaceConfigNotFound
	}

	return nil
}

func processWorkspaceConfigFilters(tx *gorm.DB, filter *config.WorkspaceConfigFilter) *gorm.DB {
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
