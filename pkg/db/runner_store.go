// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RunnerStore struct {
	IStore
}

func NewRunnerStore(store IStore) (stores.RunnerStore, error) {
	err := store.AutoMigrate(&models.Runner{})
	if err != nil {
		return nil, err
	}

	return &RunnerStore{store}, nil
}

func (s *RunnerStore) List(ctx context.Context) ([]*models.Runner, error) {
	tx := s.GetTransaction(ctx)

	runners := []*models.Runner{}
	tx = preloadRunnerEntities(tx).Find(&runners)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return runners, nil
}

func (s *RunnerStore) Find(ctx context.Context, idOrName string) (*models.Runner, error) {
	tx := s.GetTransaction(ctx)

	runner := &models.Runner{}
	tx = preloadRunnerEntities(tx).Where("id = ? OR name = ?", idOrName, idOrName).First(runner)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrRunnerNotFound
		}
		return nil, tx.Error
	}

	return runner, nil
}

func (s *RunnerStore) Save(ctx context.Context, runner *models.Runner) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Save(runner)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *RunnerStore) Delete(ctx context.Context, runner *models.Runner) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Delete(runner)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrRunnerNotFound
	}

	return nil
}

func preloadRunnerEntities(tx *gorm.DB) *gorm.DB {
	return tx.Preload(clause.Associations)
}
