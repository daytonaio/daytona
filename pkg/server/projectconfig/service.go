// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"errors"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

type IProjectConfigService interface {
	Delete(projectConfigName string) error
	Find(projectConfigName string) (*config.ProjectConfig, error)
	FindDefault(url string) (*config.ProjectConfig, error)
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

func (s *ProjectConfigService) FindDefault(url string) (*config.ProjectConfig, error) {
	url = util.CleanUpRepositoryUrl(url)

	projectConfigs, err := s.List(&config.Filter{
		Url: &url,
	})
	if err != nil {
		return nil, err
	}

	for _, pc := range projectConfigs {
		if pc.IsDefault {
			return pc, nil
		}
	}

	return nil, config.ErrProjectConfigNotFound
}

func (s *ProjectConfigService) SetDefault(projectConfigName string) error {
	projectConfig, err := s.Find(projectConfigName)
	if err != nil {
		return err
	}

	if projectConfig == nil {
		return config.ErrProjectConfigNotFound
	}

	if projectConfig.Repository == nil {
		return errors.New("project config does not have a repository")
	}

	projectConfigs, err := s.List(&config.Filter{
		Url: &projectConfig.Repository.Url,
	})
	if err != nil {
		return err
	}

	for _, pc := range projectConfigs {
		if pc.Name == projectConfigName {
			pc.IsDefault = true
		} else {
			pc.IsDefault = false
		}
		err := s.configStore.Save(pc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ProjectConfigService) Find(projectConfigName string) (*config.ProjectConfig, error) {
	return s.configStore.Find(projectConfigName)
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
	pc, err := s.Find(projectConfigName)
	if err != nil {
		return err
	}
	return s.configStore.Delete(pc)
}
