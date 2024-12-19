// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type TargetMetadataStore struct {
	IStore
}

func NewTargetMetadataStore(store IStore) (stores.TargetMetadataStore, error) {
	err := store.AutoMigrate(&models.TargetMetadata{})
	if err != nil {
		return nil, err
	}

	return &TargetMetadataStore{store}, nil
}

func (s *TargetMetadataStore) Find(ctx context.Context, targetId string) (*models.TargetMetadata, error) {
	tx := s.GetTransaction(ctx)

	targetMetadata := &models.TargetMetadata{}
	tx = tx.Where("target_id = ?", targetId).First(&targetMetadata)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrTargetMetadataNotFound
		}
		return nil, tx.Error
	}

	return targetMetadata, nil
}

func (s *TargetMetadataStore) Save(ctx context.Context, targetMetadata *models.TargetMetadata) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Save(targetMetadata)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *TargetMetadataStore) Delete(ctx context.Context, targetMetadata *models.TargetMetadata) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Delete(targetMetadata)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrTargetMetadataNotFound
	}

	return nil
}
