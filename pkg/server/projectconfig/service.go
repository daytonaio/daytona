// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/projectconfig/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

type IProjectConfigService interface {
	Save(projectConfig *config.ProjectConfig) error
	Find(filter *config.ProjectConfigFilter) (*config.ProjectConfig, error)
	List(filter *config.ProjectConfigFilter) ([]*config.ProjectConfig, error)
	SetDefault(projectConfigName string) error
	Delete(projectConfigName string, force bool) []error

	SetPrebuild(projectConfigName string, createPrebuildDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error)
	FindPrebuild(projectConfigFilter *config.ProjectConfigFilter, prebuildFilter *config.PrebuildFilter) (*dto.PrebuildDTO, error)
	ListPrebuilds(projectConfigFilter *config.ProjectConfigFilter, prebuildFilter *config.PrebuildFilter) ([]*dto.PrebuildDTO, error)
	DeletePrebuild(projectConfigName string, id string, force bool) []error

	StartRetentionPoller() error
	EnforceRetentionPolicy() error
	ProcessGitEvent(gitprovider.GitEventData) error
}

type ProjectConfigServiceConfig struct {
	PrebuildWebhookEndpoint string
	ConfigStore             config.Store
	BuildService            builds.IBuildService
	GitProviderService      gitproviders.IGitProviderService
}

type ProjectConfigService struct {
	prebuildWebhookEndpoint string
	configStore             config.Store
	buildService            builds.IBuildService
	gitProviderService      gitproviders.IGitProviderService
}

func NewProjectConfigService(config ProjectConfigServiceConfig) IProjectConfigService {

	return &ProjectConfigService{
		prebuildWebhookEndpoint: config.PrebuildWebhookEndpoint,
		configStore:             config.ConfigStore,
		buildService:            config.BuildService,
		gitProviderService:      config.GitProviderService,
	}
}

func (s *ProjectConfigService) List(filter *config.ProjectConfigFilter) ([]*config.ProjectConfig, error) {
	return s.configStore.List(filter)
}

func (s *ProjectConfigService) SetDefault(projectConfigName string) error {
	projectConfig, err := s.Find(&config.ProjectConfigFilter{
		Name: &projectConfigName,
	})
	if err != nil {
		return err
	}

	defaultProjectConfig, err := s.Find(&config.ProjectConfigFilter{
		Url:     &projectConfig.RepositoryUrl,
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

func (s *ProjectConfigService) Find(filter *config.ProjectConfigFilter) (*config.ProjectConfig, error) {
	return s.configStore.Find(filter)
}

func (s *ProjectConfigService) Save(projectConfig *config.ProjectConfig) error {
	projectConfig.RepositoryUrl = util.CleanUpRepositoryUrl(projectConfig.RepositoryUrl)

	err := s.configStore.Save(projectConfig)
	if err != nil {
		return err
	}

	return s.SetDefault(projectConfig.Name)
}

func (s *ProjectConfigService) Delete(projectConfigName string, force bool) []error {
	pc, err := s.Find(&config.ProjectConfigFilter{
		Name: &projectConfigName,
	})
	if err != nil {
		return []error{err}
	}

	// DeletePrebuild handles deleting the builds and removing the webhook
	for _, prebuild := range pc.Prebuilds {
		errs := s.DeletePrebuild(pc.Name, prebuild.Id, force)
		if len(errs) > 0 {
			return errs
		}
	}

	err = s.configStore.Delete(pc)
	if err != nil {
		return []error{err}
	}

	return nil
}
