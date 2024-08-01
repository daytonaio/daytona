// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

type IProjectConfigService interface {
	Delete(projectConfigName string) error
	Find(filter *config.Filter) (*config.ProjectConfig, error)
	List(filter *config.Filter) ([]*config.ProjectConfig, error)
	SetDefault(projectConfigName string) error
	Save(projectConfig *config.ProjectConfig) error
}

type ProjectConfigServiceConfig struct {
	ConfigStore config.Store
}

type ProjectConfigService struct {
	configStore config.Store
}

func NewConfigService(config ProjectConfigServiceConfig) IProjectConfigService {
	return &ProjectConfigService{
		configStore: config.ConfigStore,
	}
}

func (s *ProjectConfigService) List(filter *config.Filter) ([]*config.ProjectConfig, error) {
	return s.configStore.List(filter)
}

func (s *ProjectConfigService) SetDefault(projectConfigName string) error {
	projectConfig, err := s.Find(&config.Filter{
		Name: &projectConfigName,
	})
	if err != nil {
		return err
	}

	defaultProjectConfig, err := s.Find(&config.Filter{
		Url:     &projectConfig.Repository.Url,
		Default: util.Pointer(true),
	})
	if err != nil && err != config.ErrProjectConfigNotFound {
		return err
	}

	if defaultProjectConfig != nil {
		defaultProjectConfig.IsDefault = false
		err := s.configStore.Save(defaultProjectConfig)
		if err != nil {
			return err
		}
	}

	projectConfig.IsDefault = true
	return s.configStore.Save(projectConfig)
}

func (s *ProjectConfigService) Find(filter *config.Filter) (*config.ProjectConfig, error) {
	return s.configStore.Find(filter)
}

func (s *ProjectConfigService) Save(projectConfig *config.ProjectConfig) error {
	if projectConfig.Repository != nil && projectConfig.Repository.Url != "" {
		projectConfig.Repository.Url = util.CleanUpRepositoryUrl(projectConfig.Repository.Url)
	}

	err := s.configStore.Save(projectConfig)
	if err != nil {
		return err
	}

	return s.SetDefault(projectConfig.Name)
}

func (s *ProjectConfigService) Delete(projectConfigName string) error {
	pc, err := s.Find(&config.Filter{
		Name: &projectConfigName,
	})
	if err != nil {
		return err
	}
	return s.configStore.Delete(pc)
}
