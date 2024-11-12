// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targets"
)

type TargetStore struct {
	db *gorm.DB
}

func NewTargetStore(db *gorm.DB) (*TargetStore, error) {
	err := db.AutoMigrate(&models.Target{})
	if err != nil {
		return nil, err
	}

	return &TargetStore{db: db}, nil
}

func (s *TargetStore) List(filter *targets.TargetFilter) ([]*models.Target, error) {
	targets := []*models.Target{}

	tx := processTargetFilters(s.db, filter).Preload("Workspaces").Find(&targets)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return targets, nil
}

func (s *TargetStore) Find(filter *targets.TargetFilter) (*models.Target, error) {
	tg := &models.Target{}

	tx := processTargetFilters(s.db, filter).Preload("Workspaces").First(tg)
	if tx.Error != nil {
		return nil, tx.Error
	}

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, targets.ErrTargetNotFound
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
		return targets.ErrTargetNotFound
	}

	return nil
}

func processTargetFilters(tx *gorm.DB, filter *targets.TargetFilter) *gorm.DB {
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
