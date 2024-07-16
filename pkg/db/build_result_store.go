// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"github.com/daytonaio/daytona/pkg/build"
	. "github.com/daytonaio/daytona/pkg/db/dto"
	"gorm.io/gorm"
)

type BuildResultStore struct {
	db *gorm.DB
}

func NewBuildResultStore(db *gorm.DB) (*BuildResultStore, error) {
	err := db.AutoMigrate(&BuildResultDTO{})
	if err != nil {
		return nil, err
	}

	return &BuildResultStore{db: db}, nil
}

func (b *BuildResultStore) Find(hash string) (*build.BuildResult, error) {
	buildResultDTO := BuildResultDTO{}
	tx := b.db.Where("hash = ?", hash).First(&buildResultDTO)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, build.ErrBuildNotFound
		}
		return nil, tx.Error
	}

	buildResult := ToBuildResult(buildResultDTO)

	return buildResult, nil
}

func (b *BuildResultStore) List() ([]*build.BuildResult, error) {
	buildResultDTOs := []BuildResultDTO{}
	tx := b.db.Find(&buildResultDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	buildResults := []*build.BuildResult{}
	for _, buildResultDTO := range buildResultDTOs {
		buildResults = append(buildResults, ToBuildResult(buildResultDTO))
	}

	return buildResults, nil
}

func (b *BuildResultStore) Save(buildResult *build.BuildResult) error {
	buildResultDTO := ToBuildResultDTO(buildResult)
	tx := b.db.Save(&buildResultDTO)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (b *BuildResultStore) Delete(hash string) error {
	tx := b.db.Where("hash = ?", hash).Delete(&BuildResultDTO{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return build.ErrBuildNotFound
	}

	return nil
}
