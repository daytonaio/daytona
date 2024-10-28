// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/daytonaio/daytona/pkg/build"
	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/target/workspace/buildconfig"
	"gorm.io/gorm"
)

type BuildStore struct {
	db   *gorm.DB
	Lock sync.Mutex
}

func NewBuildStore(db *gorm.DB) (*BuildStore, error) {
	err := db.AutoMigrate(&BuildDTO{})
	if err != nil {
		return nil, err
	}

	return &BuildStore{db: db}, nil
}

func (b *BuildStore) Find(filter *build.Filter) (*build.Build, error) {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	buildDTO := BuildDTO{}
	tx := processBuildFilters(b.db, filter).First(&buildDTO)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, build.ErrBuildNotFound
		}
		return nil, tx.Error
	}

	build := ToBuild(buildDTO)

	return build, nil
}

func (b *BuildStore) List(filter *build.Filter) ([]*build.Build, error) {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	buildDTOs := []BuildDTO{}
	tx := processBuildFilters(b.db, filter).Find(&buildDTOs)

	if tx.Error != nil {
		return nil, tx.Error
	}

	builds := []*build.Build{}
	for _, buildDTO := range buildDTOs {
		builds = append(builds, ToBuild(buildDTO))
	}

	return builds, nil
}

func (b *BuildStore) Save(build *build.Build) error {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	buildDTO := ToBuildDTO(build)
	tx := b.db.Save(&buildDTO)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (b *BuildStore) Delete(id string) error {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	tx := b.db.Where("id = ?", id).Delete(&BuildDTO{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return build.ErrBuildNotFound
	}

	return nil
}

func processBuildFilters(tx *gorm.DB, filter *build.Filter) *gorm.DB {
	if filter != nil {
		if filter.Id != nil {
			tx = tx.Where("id = ?", *filter.Id)
		}
		if filter.States != nil && len(*filter.States) > 0 {
			placeholders := strings.Repeat("?,", len(*filter.States))
			placeholders = placeholders[:len(placeholders)-1]

			tx = tx.Where(fmt.Sprintf("state IN (%s)", placeholders), filter.StatesToInterface()...)
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
		if filter.BuildConfig != nil && *filter.BuildConfig != (buildconfig.BuildConfig{}) {
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
