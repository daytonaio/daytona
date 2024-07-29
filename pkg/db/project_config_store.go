// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

type ProjectConfigStore struct {
	db *gorm.DB
}

func NewProjectConfigStore(db *gorm.DB) (*ProjectConfigStore, error) {
	err := db.AutoMigrate(&CreateProjectConfigDTO{})
	if err != nil {
		return nil, err
	}

	return &ProjectConfigStore{db: db}, nil
}

func (s *ProjectConfigStore) List() ([]*config.ProjectConfig, error) {
	projectConfigsDTOs := []CreateProjectConfigDTO{}
	tx := s.db.Find(&projectConfigsDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	projectConfigs := []*config.ProjectConfig{}
	for _, projectConfigDTO := range projectConfigsDTOs {
		projectConfigs = append(projectConfigs, ToProjectConfig(projectConfigDTO))
	}

	return projectConfigs, nil
}

func (s *ProjectConfigStore) Find(projectConfigName string) (*config.ProjectConfig, error) {
	projectConfigDTO := CreateProjectConfigDTO{}
	tx := s.db.Where("name = ?", projectConfigName).First(&projectConfigDTO)
	if tx.Error != nil {
		if IsRecordNotFound(tx.Error) {
			return nil, config.ErrProjectConfigNotFound
		}
		return nil, tx.Error
	}

	return ToProjectConfig(projectConfigDTO), nil
}

func (s *ProjectConfigStore) Save(projectConfig *config.ProjectConfig) error {
	tx := s.db.Save(ToCreateProjectConfigDTO(projectConfig))
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (s *ProjectConfigStore) Delete(projectConfig *config.ProjectConfig) error {
	tx := s.db.Delete(ToCreateProjectConfigDTO(projectConfig))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return config.ErrProjectConfigNotFound
	}

	return nil
}
