// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type WorkspaceTemplateStore struct {
	db *gorm.DB
}

func NewWorkspaceTemplateStore(db *gorm.DB) (*WorkspaceTemplateStore, error) {
	err := db.AutoMigrate(&models.WorkspaceTemplate{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceTemplateStore{db: db}, nil
}

func (s *WorkspaceTemplateStore) List(filter *stores.WorkspaceTemplateFilter) ([]*models.WorkspaceTemplate, error) {
	workspaceTemplates := []*models.WorkspaceTemplate{}
	tx := processWorkspaceTemplateFilters(s.db, filter).Find(&workspaceTemplates)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return workspaceTemplates, nil
}

func (s *WorkspaceTemplateStore) Find(filter *stores.WorkspaceTemplateFilter) (*models.WorkspaceTemplate, error) {
	workspaceTemplate := &models.WorkspaceTemplate{}
	tx := processWorkspaceTemplateFilters(s.db, filter).First(workspaceTemplate)

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrWorkspaceTemplateNotFound
		}
		return nil, tx.Error
	}

	return workspaceTemplate, nil
}

func (s *WorkspaceTemplateStore) Save(workspaceTemplate *models.WorkspaceTemplate) error {
	tx := s.db.Save(workspaceTemplate)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *WorkspaceTemplateStore) Delete(workspaceTemplate *models.WorkspaceTemplate) error {
	tx := s.db.Delete(workspaceTemplate)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrWorkspaceTemplateNotFound
	}

	return nil
}

func processWorkspaceTemplateFilters(tx *gorm.DB, filter *stores.WorkspaceTemplateFilter) *gorm.DB {
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
