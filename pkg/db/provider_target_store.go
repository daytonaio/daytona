// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/provider"
)

type ProviderTargetStore struct {
	db *gorm.DB
}

func NewProviderTargetStore(db *gorm.DB) (*ProviderTargetStore, error) {
	err := db.AutoMigrate(&ProviderTargetDTO{})
	if err != nil {
		return nil, err
	}

	return &ProviderTargetStore{db: db}, nil
}

func (s *ProviderTargetStore) List(filter *provider.TargetFilter) ([]*provider.ProviderTarget, error) {
	targetDTOs := []ProviderTargetDTO{}
	tx := processTargetFilters(s.db, filter).Find(&targetDTOs)

	if tx.Error != nil {
		return nil, tx.Error
	}

	targets := []*provider.ProviderTarget{}
	for _, targetDTO := range targetDTOs {
		targets = append(targets, ToProviderTarget(targetDTO))
	}

	return targets, nil
}

func (s *ProviderTargetStore) Find(filter *provider.TargetFilter) (*provider.ProviderTarget, error) {
	targetDTO := ProviderTargetDTO{}
	tx := processTargetFilters(s.db, filter).First(&targetDTO)

	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, provider.ErrTargetNotFound
		}
		return nil, tx.Error
	}

	return ToProviderTarget(targetDTO), nil
}

func (s *ProviderTargetStore) Save(target *provider.ProviderTarget) error {
	tx := s.db.Save(ToProviderTargetDTO(target))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *ProviderTargetStore) Delete(target *provider.ProviderTarget) error {
	tx := s.db.Delete(ToProviderTargetDTO(target))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return provider.ErrTargetNotFound
	}

	return nil
}

func processTargetFilters(tx *gorm.DB, filter *provider.TargetFilter) *gorm.DB {
	if filter != nil {
		if filter.Name != nil {
			tx = tx.Where("name = ?", *filter.Name)
		}
		if filter.Default != nil {
			tx = tx.Where("is_default = ?", *filter.Default)
		}
	}

	return tx
}
