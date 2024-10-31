// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/internal/util"
	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/target"
)

type TargetStore struct {
	db *gorm.DB
}

func NewTargetStore(db *gorm.DB) (*TargetStore, error) {
	err := db.AutoMigrate(&TargetDTO{})
	if err != nil {
		return nil, err
	}

	return &TargetStore{db: db}, nil
}

func (s *TargetStore) List(filter *target.TargetFilter) ([]*target.Target, error) {
	targetDTOs := []TargetDTO{}
	tx := processTargetFilters(s.db, filter).Find(&targetDTOs)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return util.ArrayMap(targetDTOs, func(targetDTOs TargetDTO) *target.Target {
		return ToTarget(targetDTOs)
	}), nil
}

func (s *TargetStore) Find(filter *target.TargetFilter) (*target.Target, error) {
	targetDTO := TargetDTO{}
	tx := processTargetFilters(s.db, filter).First(&targetDTO)

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, target.ErrTargetNotFound
		}
		return nil, tx.Error
	}

	return ToTarget(targetDTO), nil
}

func (s *TargetStore) Save(target *target.Target) error {
	tx := s.db.Save(ToTargetDTO(target))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *TargetStore) Delete(t *target.Target) error {
	tx := s.db.Delete(ToTargetDTO(t))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return target.ErrTargetNotFound
	}

	return nil
}

func processTargetFilters(tx *gorm.DB, filter *target.TargetFilter) *gorm.DB {
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
