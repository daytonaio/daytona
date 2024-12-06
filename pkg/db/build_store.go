// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"gorm.io/gorm"
)

type BuildStore struct {
	IStore
	Lock sync.Mutex
}

func NewBuildStore(store IStore) (stores.BuildStore, error) {
	err := store.AutoMigrate(&models.Build{})
	if err != nil {
		return nil, err
	}

	return &BuildStore{store, sync.Mutex{}}, nil
}

func (b *BuildStore) Find(ctx context.Context, filter *stores.BuildFilter) (*models.Build, error) {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	tx := b.GetTransaction(ctx)

	build := &models.Build{}
	tx = preloadBuildEntities(processBuildFilters(tx, filter)).First(build)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, stores.ErrBuildNotFound
		}
		return nil, tx.Error
	}

	return build, nil
}

func (b *BuildStore) List(ctx context.Context, filter *stores.BuildFilter) ([]*models.Build, error) {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	tx := b.GetTransaction(ctx)

	builds := []*models.Build{}
	tx = preloadBuildEntities(processBuildFilters(tx, filter)).Find(&builds)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return builds, nil
}

func (b *BuildStore) Save(ctx context.Context, build *models.Build) error {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	tx := b.GetTransaction(ctx)

	tx = tx.Save(build)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (b *BuildStore) Delete(ctx context.Context, id string) error {
	b.Lock.Lock()
	defer b.Lock.Unlock()
	tx := b.GetTransaction(ctx)

	tx = tx.Where("id = ?", id).Delete(&models.Build{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return stores.ErrBuildNotFound
	}

	return nil
}

func preloadBuildEntities(tx *gorm.DB) *gorm.DB {
	return tx.Preload("LastJob", preloadLastJob)
}

func processBuildFilters(tx *gorm.DB, filter *stores.BuildFilter) *gorm.DB {
	if filter != nil {
		if filter.Id != nil {
			tx = tx.Where("id = ?", *filter.Id)
		}
		if filter.PrebuildIds != nil && len(*filter.PrebuildIds) > 0 {
			placeholders := strings.Repeat("?,", len(*filter.PrebuildIds))
			placeholders = placeholders[:len(placeholders)-1]

			tx = tx.Where(fmt.Sprintf("prebuild_id IN (%s)", placeholders), stringsToInterface(*filter.PrebuildIds)...)
		}
		if filter.GetNewest != nil && *filter.GetNewest {
			tx = tx.Order("created_at desc").Limit(1)
		}
		// Skip filtering when an automatic build config is provided
		if filter.BuildConfig != nil && *filter.BuildConfig != (models.BuildConfig{}) {
			buildConfigJSON, err := json.Marshal(filter.BuildConfig)
			if err == nil {
				tx = tx.Where("build_config = ?", string(buildConfigJSON))
			}
		}
		if filter.RepositoryUrl != nil {
			tx = tx.Where("json_extract(repository, '$.url') = ?", *filter.RepositoryUrl)
		}
		if filter.Branch != nil {
			tx = tx.Where("json_extract(repository, '$.branch') = ?", *filter.Branch)
		}
		if filter.EnvVars != nil && len(*filter.EnvVars) > 0 {
			envVarsJSON, err := json.Marshal(filter.EnvVars)
			if err == nil {
				tx = tx.Where("env_vars = ?", string(envVarsJSON))
			}
		}
	}
	return tx
}

func stringsToInterface(slice []string) []interface{} {
	args := make([]interface{}, len(slice))
	for i, v := range slice {
		args[i] = v
	}
	return args
}
