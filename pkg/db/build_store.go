// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"sync"

	"github.com/daytonaio/daytona/pkg/build"
	. "github.com/daytonaio/daytona/pkg/db/dto"
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

func (b *BuildStore) Find(hash string) (*build.Build, error) {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	buildDTO := BuildDTO{}
	tx := b.db.Where("hash = ?", hash).First(&buildDTO)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, build.ErrBuildNotFound
		}
		return nil, tx.Error
	}

	build := ToBuild(buildDTO)

	return build, nil
}

func (b *BuildStore) List(filter *build.BuildFilter) ([]*build.Build, error) {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	buildDTOs := []BuildDTO{}
	tx := b.db

	if filter != nil {
		if filter.State != nil {
			tx = b.db.Where("state = ?", *filter.State)
		}
	}
	tx = tx.Find(&buildDTOs)

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

func (b *BuildStore) Delete(hash string) error {
	b.Lock.Lock()
	defer b.Lock.Unlock()

	tx := b.db.Where("hash = ?", hash).Delete(&BuildDTO{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return build.ErrBuildNotFound
	}

	return nil
}
