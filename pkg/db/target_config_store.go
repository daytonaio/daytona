// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/internal/util"
	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/target/config"
)

type TargetConfigStore struct {
	db *gorm.DB
}

func NewTargetConfigStore(db *gorm.DB) (*TargetConfigStore, error) {
	err := db.AutoMigrate(&TargetConfigDTO{})
	if err != nil {
		return nil, err
	}

	return &TargetConfigStore{db: db}, nil
}

func (s *TargetConfigStore) List(filter *config.TargetConfigFilter) ([]*config.TargetConfig, error) {
	targetConfigDTOs := []TargetConfigDTO{}
	tx := processTargetConfigFilters(s.db, filter).Find(&targetConfigDTOs)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return util.ArrayMap(targetConfigDTOs, func(targetConfigDTO TargetConfigDTO) *config.TargetConfig {
		return ToTargetConfig(targetConfigDTO)
	}), nil
}

func (s *TargetConfigStore) Find(filter *config.TargetConfigFilter) (*config.TargetConfig, error) {
	targetConfigDTO := TargetConfigDTO{}
	tx := processTargetConfigFilters(s.db, filter).First(&targetConfigDTO)

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, config.ErrTargetConfigNotFound
		}
		return nil, tx.Error
	}

	return ToTargetConfig(targetConfigDTO), nil
}

func (s *TargetConfigStore) Save(targetConfig *config.TargetConfig) error {
	tx := s.db.Save(ToTargetConfigDTO(targetConfig))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *TargetConfigStore) Delete(targetConfig *config.TargetConfig) error {
	tx := s.db.Delete(ToTargetConfigDTO(targetConfig))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return config.ErrTargetConfigNotFound
	}

	return nil
}

func processTargetConfigFilters(tx *gorm.DB, filter *config.TargetConfigFilter) *gorm.DB {
	if filter != nil {
		if filter.Name != nil {
			tx = tx.Where("name = ?", *filter.Name)
		}
	}

	return tx
}
