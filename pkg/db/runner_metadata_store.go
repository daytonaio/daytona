// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type RunnerMetadataStore struct {
	IStore
}

func NewRunnerMetadataStore(store IStore) (stores.RunnerMetadataStore, error) {
	err := store.AutoMigrate(&models.RunnerMetadata{})
	if err != nil {
		return nil, err
	}

	return &RunnerMetadataStore{store}, nil
}

func (s *RunnerMetadataStore) List(ctx context.Context) ([]*models.RunnerMetadata, error) {
	tx := s.GetTransaction(ctx)

	var runnerMetadata []*models.RunnerMetadata
	tx = tx.Find(&runnerMetadata)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return runnerMetadata, nil
}

func (s *RunnerMetadataStore) Find(ctx context.Context, runnerId string) (*models.RunnerMetadata, error) {
	tx := s.GetTransaction(ctx)

	runnerMetadata := &models.RunnerMetadata{}
	tx = tx.Where("runner_id = ?", runnerId).First(&runnerMetadata)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrRunnerMetadataNotFound
		}
		return nil, tx.Error
	}

	return runnerMetadata, nil
}

func (s *RunnerMetadataStore) Save(ctx context.Context, runnerMetadata *models.RunnerMetadata) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Save(runnerMetadata)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *RunnerMetadataStore) Delete(ctx context.Context, runnerMetadata *models.RunnerMetadata) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Delete(runnerMetadata)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrRunnerMetadataNotFound
	}

	return nil
}
