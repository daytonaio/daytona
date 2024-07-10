// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/prebuild"
	"gorm.io/gorm"
)

type PrebuildStore struct {
	db *gorm.DB
}

func NewPrebuildStore(db *gorm.DB) (*PrebuildStore, error) {
	err := db.AutoMigrate(&PrebuildDTO{})
	if err != nil {
		return nil, err
	}

	return &PrebuildStore{db: db}, nil
}

func (s *PrebuildStore) Find(key string) (*prebuild.Prebuild, error) {
	prebuildConfigDTO := PrebuildDTO{}
	tx := s.db.Where("key = ?", key).First(&prebuildConfigDTO)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, prebuild.ErrPrebuildNotFound
		}
		return nil, tx.Error
	}

	prebuildConfig := ToPrebuild(prebuildConfigDTO)

	return prebuildConfig, nil
}

func (s *PrebuildStore) Save(p *prebuild.Prebuild) error {
	prebuildConfigDTO := ToPrebuildDTO(p)
	tx := s.db.Save(&prebuildConfigDTO)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *PrebuildStore) List(filter *prebuild.PrebuildFilter) ([]*prebuild.Prebuild, error) {
	prebuildConfigDTOs := []*PrebuildDTO{}
	tx := s.db

	if filter != nil {
		if filter.ProjectConfigName != "" {
			tx = s.db.Where("json_extract(projectConfig, '$.name') = ?", filter.ProjectConfigName)
		}
	}

	tx.Find(&prebuildConfigDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	prebuildConfigs := []*prebuild.Prebuild{}
	for _, prebuildConfigDTO := range prebuildConfigDTOs {
		prebuildConfigs = append(prebuildConfigs, ToPrebuild(*prebuildConfigDTO))
	}

	return prebuildConfigs, nil
}

func (s *PrebuildStore) Delete(p *prebuild.Prebuild) error {
	tx := s.db.Where("key = ?", p.Key).Delete(&PrebuildDTO{})
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return prebuild.ErrPrebuildNotFound
	}

	return nil
}
