// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfig/dto"
	"github.com/daytonaio/daytona/pkg/target/workspace/config"
)

type IWorkspaceConfigService interface {
	Save(workspaceConfig *config.WorkspaceConfig) error
	Find(filter *config.WorkspaceConfigFilter) (*config.WorkspaceConfig, error)
	List(filter *config.WorkspaceConfigFilter) ([]*config.WorkspaceConfig, error)
	SetDefault(workspaceConfigName string) error
	Delete(workspaceConfigName string, force bool) []error

	SetPrebuild(workspaceConfigName string, createPrebuildDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error)
	FindPrebuild(workspaceConfigFilter *config.WorkspaceConfigFilter, prebuildFilter *config.PrebuildFilter) (*dto.PrebuildDTO, error)
	ListPrebuilds(workspaceConfigFilter *config.WorkspaceConfigFilter, prebuildFilter *config.PrebuildFilter) ([]*dto.PrebuildDTO, error)
	DeletePrebuild(workspaceConfigName string, id string, force bool) []error

	StartRetentionPoller() error
	EnforceRetentionPolicy() error
	ProcessGitEvent(gitprovider.GitEventData) error
}

type WorkspaceConfigServiceConfig struct {
	PrebuildWebhookEndpoint string
	ConfigStore             config.Store
	BuildService            builds.IBuildService
	GitProviderService      gitproviders.IGitProviderService
}

type WorkspaceConfigService struct {
	prebuildWebhookEndpoint string
	configStore             config.Store
	buildService            builds.IBuildService
	gitProviderService      gitproviders.IGitProviderService
}

func NewWorkspaceConfigService(config WorkspaceConfigServiceConfig) IWorkspaceConfigService {
	return &WorkspaceConfigService{
		prebuildWebhookEndpoint: config.PrebuildWebhookEndpoint,
		configStore:             config.ConfigStore,
		buildService:            config.BuildService,
		gitProviderService:      config.GitProviderService,
	}
}

func (s *WorkspaceConfigService) List(filter *config.WorkspaceConfigFilter) ([]*config.WorkspaceConfig, error) {
	return s.configStore.List(filter)
}

func (s *WorkspaceConfigService) SetDefault(workspaceConfigName string) error {
	workspaceConfig, err := s.Find(&config.WorkspaceConfigFilter{
		Name: &workspaceConfigName,
	})
	if err != nil {
		return err
	}

	defaultWorkspaceConfig, err := s.Find(&config.WorkspaceConfigFilter{
		Url:     &workspaceConfig.RepositoryUrl,
		Default: util.Pointer(true),
	})
	if err != nil && err != config.ErrWorkspaceConfigNotFound {
		return err
	}

	if defaultWorkspaceConfig != nil {
		defaultWorkspaceConfig.IsDefault = false
		err := s.configStore.Save(defaultWorkspaceConfig)
		if err != nil {
			return err
		}
	}

	workspaceConfig.IsDefault = true
	return s.configStore.Save(workspaceConfig)
}

func (s *WorkspaceConfigService) Find(filter *config.WorkspaceConfigFilter) (*config.WorkspaceConfig, error) {
	if filter != nil && filter.Url != nil {
		cleanedUrl := util.CleanUpRepositoryUrl(*filter.Url)
		if !strings.HasSuffix(cleanedUrl, ".git") {
			cleanedUrl = cleanedUrl + ".git"
		}
		filter.Url = util.Pointer(cleanedUrl)
	}
	return s.configStore.Find(filter)
}

func (s *WorkspaceConfigService) Save(workspaceConfig *config.WorkspaceConfig) error {
	workspaceConfig.RepositoryUrl = util.CleanUpRepositoryUrl(workspaceConfig.RepositoryUrl)

	err := s.configStore.Save(workspaceConfig)
	if err != nil {
		return err
	}

	return s.SetDefault(workspaceConfig.Name)
}

func (s *WorkspaceConfigService) Delete(workspaceConfigName string, force bool) []error {
	wc, err := s.Find(&config.WorkspaceConfigFilter{
		Name: &workspaceConfigName,
	})
	if err != nil {
		return []error{err}
	}

	// DeletePrebuild handles deleting the builds and removing the webhook
	for _, prebuild := range wc.Prebuilds {
		errs := s.DeletePrebuild(wc.Name, prebuild.Id, force)
		if len(errs) > 0 {
			return errs
		}
	}

	err = s.configStore.Delete(wc)
	if err != nil {
		return []error{err}
	}

	return nil
}
