// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
)

type TargetConfigStore struct {
	db *gorm.DB
}

func NewTargetConfigStore(db *gorm.DB) (*TargetConfigStore, error) {
	err := db.AutoMigrate(&models.TargetConfig{})
	if err != nil {
		return nil, err
	}

	return &TargetConfigStore{db: db}, nil
}

func (s *TargetConfigStore) List(filter *targetconfigs.TargetConfigFilter) ([]*models.TargetConfig, error) {
	targetConfigs := []*models.TargetConfig{}
	tx := processTargetConfigFilters(s.db, filter).Find(&targetConfigs)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return targetConfigs, nil
}

func (s *TargetConfigStore) Find(filter *targetconfigs.TargetConfigFilter) (*models.TargetConfig, error) {
	targetConfig := &models.TargetConfig{}
	tx := processTargetConfigFilters(s.db, filter).First(targetConfig)

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, targetconfigs.ErrTargetConfigNotFound
		}
		return nil, tx.Error
	}

	return targetConfig, nil
}

func (s *TargetConfigStore) Save(targetConfig *models.TargetConfig) error {
	tx := s.db.Save(targetConfig)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *TargetConfigStore) Delete(targetConfig *models.TargetConfig) error {
	tx := s.db.Delete(targetConfig)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return targetconfigs.ErrTargetConfigNotFound
	}

	return nil
}

func processTargetConfigFilters(tx *gorm.DB, filter *targetconfigs.TargetConfigFilter) *gorm.DB {
	if filter != nil {
		if filter.Name != nil {
			tx = tx.Where("name = ?", *filter.Name)
		}
	}

	return tx
}
