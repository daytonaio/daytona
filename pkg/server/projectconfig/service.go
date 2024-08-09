// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/server/projectconfig/prebuild"
	"github.com/daytonaio/daytona/pkg/server/projectconfig/prebuild/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	prebuild_cnf "github.com/daytonaio/daytona/pkg/workspace/project/config/prebuild"
)

type IProjectConfigService interface {
	Delete(projectConfigName string) error
	Find(filter *config.Filter) (*config.ProjectConfig, error)
	List(filter *config.Filter) ([]*config.ProjectConfig, error)
	SetDefault(projectConfigName string) error
	Save(projectConfig *config.ProjectConfig) error
	SetPrebuild(dto.CreatePrebuildDTO) error
	FindPrebuild(projectConfigName, id string) (*prebuild_cnf.PrebuildConfig, error)
	ListPrebuilds(*config.PrebuildFilter) ([]*dto.PrebuildDTO, error)
	DeletePrebuild(projectConfigName string, prebuild *prebuild_cnf.PrebuildConfig) error
}

type ProjectConfigServiceConfig struct {
	ConfigStore     config.Store
	BuildService    builds.IBuildService
	PrebuildService prebuild.IPrebuildService
}

type ProjectConfigService struct {
	configStore     config.Store
	buildService    builds.IBuildService
	prebuildService prebuild.IPrebuildService
}

func NewProjectConfigService(config ProjectConfigServiceConfig) IProjectConfigService {
	prebuildService := prebuild.NewPrebuildService(config.ConfigStore)

	return &ProjectConfigService{
		configStore:     config.ConfigStore,
		buildService:    config.BuildService,
		prebuildService: prebuildService,
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

func (s *ProjectConfigService) SetPrebuild(createPrebuildDto dto.CreatePrebuildDTO) error {
	err := s.prebuildService.Set(createPrebuildDto)
	if err != nil {
		return err
	}

	if createPrebuildDto.RunAtInit {
		projectConfig, err := s.Find(&config.Filter{
			Name: &createPrebuildDto.ProjectConfigName,
		})
		if err != nil {
			return err
		}

		build := &build.Build{
			ProjectConfig: *projectConfig,
		}

		err = s.buildService.Create(build)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ProjectConfigService) FindPrebuild(projectConfigName, id string) (*prebuild_cnf.PrebuildConfig, error) {
	return nil, nil
}

func (s *ProjectConfigService) ListPrebuilds(*config.PrebuildFilter) ([]*dto.PrebuildDTO, error) {
	return nil, nil
}

func (s *ProjectConfigService) DeletePrebuild(projectConfigName string, prebuild *prebuild_cnf.PrebuildConfig) error {
	return nil
}
