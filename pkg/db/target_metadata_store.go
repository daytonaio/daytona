// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type TargetMetadataStore struct {
	db *gorm.DB
}

func NewTargetMetadataStore(db *gorm.DB) (*TargetMetadataStore, error) {
	err := db.AutoMigrate(&models.TargetMetadata{})
	if err != nil {
		return nil, err
	}

	return &TargetMetadataStore{db: db}, nil
}

func (s *TargetMetadataStore) Find(filter *stores.TargetMetadataFilter) (*models.TargetMetadata, error) {
	targetMetadata := &models.TargetMetadata{}
	tx := processTargetMetadataFilters(s.db, filter).First(&targetMetadata)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrTargetMetadataNotFound
		}
		return nil, tx.Error
	}

	return targetMetadata, nil
}

func (s *TargetMetadataStore) Save(targetMetadata *models.TargetMetadata) error {
	tx := s.db.Save(targetMetadata)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *TargetMetadataStore) Delete(targetMetadata *models.TargetMetadata) error {
	tx := s.db.Delete(targetMetadata)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrTargetMetadataNotFound
	}

	return nil
}

func processTargetMetadataFilters(tx *gorm.DB, filter *stores.TargetMetadataFilter) *gorm.DB {
	if filter != nil {
		if filter.Id != nil {
			tx = tx.Where("id = ?", *filter.Id)
		}
		if filter.TargetId != nil {
			tx = tx.Where("target_id = ?", *filter.TargetId)
		}
	}
	return tx
}
