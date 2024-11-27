// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"gorm.io/gorm"
)

type JobStore struct {
	db *gorm.DB
}

func NewJobStore(db *gorm.DB) (*JobStore, error) {
	err := db.AutoMigrate(&models.Job{})
	if err != nil {
		return nil, err
	}

	return &JobStore{db: db}, nil
}

func (s *JobStore) List(filter *stores.JobFilter) ([]*models.Job, error) {
	jobs := []*models.Job{}
	tx := processJobFilters(s.db, filter).Find(&jobs)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return jobs, nil
}

func (s *JobStore) Find(filter *stores.JobFilter) (*models.Job, error) {
	job := &models.Job{}
	tx := processJobFilters(s.db, filter).First(&job)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, stores.ErrJobNotFound
		}
		return nil, tx.Error
	}

	return job, nil

}

func (s *JobStore) Save(job *models.Job) error {
	tx := s.db.Save(job)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *JobStore) Delete(job *models.Job) error {
	tx := s.db.Delete(job)
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
