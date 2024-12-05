// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type WorkspaceTemplateStore struct {
	Store
}

func NewWorkspaceTemplateStore(store Store) (stores.WorkspaceTemplateStore, error) {
	err := store.db.AutoMigrate(&models.WorkspaceTemplate{})
	if err != nil {
		return nil, err
	}

	return &WorkspaceTemplateStore{store}, nil
}

func (s *WorkspaceTemplateStore) List(ctx context.Context, filter *stores.WorkspaceTemplateFilter) ([]*models.WorkspaceTemplate, error) {
	tx := s.getTransaction(ctx)

	workspaceTemplates := []*models.WorkspaceTemplate{}
	tx = processWorkspaceTemplateFilters(tx, filter).Find(&workspaceTemplates)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return workspaceTemplates, nil
}

func (s *WorkspaceTemplateStore) Find(ctx context.Context, filter *stores.WorkspaceTemplateFilter) (*models.WorkspaceTemplate, error) {
	tx := s.getTransaction(ctx)

	workspaceTemplate := &models.WorkspaceTemplate{}
	tx = processWorkspaceTemplateFilters(tx, filter).First(workspaceTemplate)

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrWorkspaceTemplateNotFound
		}
		return nil, tx.Error
	}

	return workspaceTemplate, nil
}

func (s *WorkspaceTemplateStore) Save(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate) error {
	tx := s.getTransaction(ctx)

	tx = tx.Save(workspaceTemplate)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *WorkspaceTemplateStore) Delete(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate) error {
	tx := s.getTransaction(ctx)

	tx = tx.Delete(workspaceTemplate)
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
