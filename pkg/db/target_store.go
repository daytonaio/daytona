// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type TargetStore struct {
	IStore
}

func NewTargetStore(store IStore) (stores.TargetStore, error) {
	err := store.AutoMigrate(&models.Target{})
	if err != nil {
		return nil, err
	}

	return &TargetStore{store}, nil
}

func (s *TargetStore) List(ctx context.Context, filter *stores.TargetFilter) ([]*models.Target, error) {
	tx := s.GetTransaction(ctx)

	targets := []*models.Target{}

	tx = preloadTargetEntities(processTargetFilters(tx, filter)).Find(&targets)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return targets, nil
}

func (s *TargetStore) Find(ctx context.Context, filter *stores.TargetFilter) (*models.Target, error) {
	tx := s.GetTransaction(ctx)

	tg := &models.Target{}

	tx = preloadTargetEntities(processTargetFilters(tx, filter)).First(tg)
	if tx.Error != nil {
		return nil, tx.Error
	}

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrTargetNotFound
		}
		return nil, tx.Error
	}
	return tg, nil
}

func (s *TargetStore) Save(ctx context.Context, target *models.Target) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Save(target)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *TargetStore) Delete(ctx context.Context, t *models.Target) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Delete(t)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrTargetNotFound
	}

	return nil
}

func processTargetFilters(tx *gorm.DB, filter *stores.TargetFilter) *gorm.DB {
	if filter != nil {
		if filter.IdOrName != nil {
			tx = tx.Where("id = ? OR name = ?", *filter.IdOrName, *filter.IdOrName)
		}
		if filter.Default != nil {
			tx = tx.Where("is_default = ?", *filter.Default)
		}
	}

	return tx
}

func preloadTargetEntities(tx *gorm.DB) *gorm.DB {
	return tx.Preload(clause.Associations).Preload("Workspaces.LastJob").Preload("LastJob")
}
