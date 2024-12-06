// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"gorm.io/gorm"
)

type JobStore struct {
	IStore
}

func NewJobStore(store IStore) (stores.JobStore, error) {
	err := store.AutoMigrate(&models.Job{})
	if err != nil {
		return nil, err
	}

	return &JobStore{store}, nil
}

func (s *JobStore) List(ctx context.Context, filter *stores.JobFilter) ([]*models.Job, error) {
	tx := s.GetTransaction(ctx)

	jobs := []*models.Job{}
	tx = processJobFilters(tx, filter).Find(&jobs)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return jobs, nil
}

func (s *JobStore) Find(ctx context.Context, filter *stores.JobFilter) (*models.Job, error) {
	tx := s.GetTransaction(ctx)

	job := &models.Job{}
	tx = processJobFilters(tx, filter).First(&job)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrJobNotFound
		}
		return nil, tx.Error
	}

	return job, nil

}

func (s *JobStore) Save(ctx context.Context, job *models.Job) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Save(job)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *JobStore) Delete(ctx context.Context, job *models.Job) error {
	tx := s.GetTransaction(ctx)

	tx = tx.Delete(job)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrJobNotFound
	}

	return nil
}

func processJobFilters(tx *gorm.DB, filter *stores.JobFilter) *gorm.DB {
	if filter != nil {
		if filter.Id != nil {
			tx = tx.Where("id = ?", *filter.Id)
		}
		if filter.ResourceType != nil {
			tx = tx.Where("resource_type = ?", *filter.ResourceType)
		}
		if filter.ResourceId != nil {
			tx = tx.Where("resource_id = ?", *filter.ResourceId)
		}
		if filter.States != nil && len(*filter.States) > 0 {
			placeholders := strings.Repeat("?,", len(*filter.States))
			placeholders = placeholders[:len(placeholders)-1]

			tx = tx.Where(fmt.Sprintf("state IN (%s)", placeholders), filter.StatesToInterface()...)
		}
		if filter.Actions != nil && len(*filter.Actions) > 0 {
			placeholders := strings.Repeat("?,", len(*filter.Actions))
			placeholders = placeholders[:len(placeholders)-1]

			tx = tx.Where(fmt.Sprintf("action IN (%s)", placeholders), filter.ActionsToInterface()...)
		}
	}
	return tx
}

func preloadLastJob(tx *gorm.DB) *gorm.DB {
	return tx.Where("updated_at IN (?)",
		tx.Model(&models.Job{}).
			Select("MAX(updated_at)").
			Group("resource_id"),
	)
}
