// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type TargetStore struct {
	db *gorm.DB
}

func NewTargetStore(db *gorm.DB) (stores.TargetStore, error) {
	err := db.AutoMigrate(&models.Target{})
	if err != nil {
		return nil, err
	}

	return &TargetStore{db: db}, nil
}

func (s *TargetStore) List(filter *stores.TargetFilter) ([]*models.Target, error) {
	targets := []*models.Target{}

	tx := preloadTargetEntities(processTargetFilters(s.db, filter)).Find(&targets)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return targets, nil
}

func (s *TargetStore) Find(filter *stores.TargetFilter) (*models.Target, error) {
	tg := &models.Target{}

	tx := preloadTargetEntities(processTargetFilters(s.db, filter)).First(tg)
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

func (s *TargetStore) Save(target *models.Target) error {
	tx := s.db.Save(target)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *TargetStore) Delete(t *models.Target) error {
	tx := s.db.Delete(t)
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
	return tx.Preload(clause.Associations).Preload("Workspaces.LastJob", preloadLastJob).Preload("LastJob", preloadLastJob)
}
