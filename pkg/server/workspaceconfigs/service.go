// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfigs

import (
	"context"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

type WorkspaceConfigServiceConfig struct {
	PrebuildWebhookEndpoint string
	ConfigStore             stores.WorkspaceConfigStore

	FindNewestBuild           func(ctx context.Context, prebuildId string) (*models.Build, error)
	ListPublishedBuilds       func(ctx context.Context) ([]*models.Build, error)
	CreateBuild               func(ctx context.Context, wc *models.WorkspaceConfig, repo *gitprovider.GitRepository, prebuildId string) error
	DeleteBuilds              func(ctx context.Context, id, prebuildId *string, force bool) []error
	GetRepositoryContext      func(ctx context.Context, url string) (repo *gitprovider.GitRepository, gitProviderId string, err error)
	FindPrebuildWebhook       func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error)
	UnregisterPrebuildWebhook func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, id string) error
	RegisterPrebuildWebhook   func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error)
	GetCommitsRange           func(ctx context.Context, repo *gitprovider.GitRepository, initialSha string, currentSha string) (int, error)
}

type WorkspaceConfigService struct {
	prebuildWebhookEndpoint string
	configStore             stores.WorkspaceConfigStore

	findNewestBuild           func(ctx context.Context, prebuildId string) (*models.Build, error)
	listPublishedBuilds       func(ctx context.Context) ([]*models.Build, error)
	createBuild               func(ctx context.Context, wc *models.WorkspaceConfig, repo *gitprovider.GitRepository, prebuildId string) error
	deleteBuilds              func(ctx context.Context, id, prebuildId *string, force bool) []error
	getRepositoryContext      func(ctx context.Context, url string) (repo *gitprovider.GitRepository, gitProviderId string, err error)
	getCommitsRange           func(ctx context.Context, repo *gitprovider.GitRepository, initialSha string, currentSha string) (int, error)
	findPrebuildWebhook       func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error)
	unregisterPrebuildWebhook func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, id string) error
	registerPrebuildWebhook   func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error)
}

func NewWorkspaceConfigService(config WorkspaceConfigServiceConfig) services.IWorkspaceConfigService {
	return &WorkspaceConfigService{
		prebuildWebhookEndpoint:   config.PrebuildWebhookEndpoint,
		configStore:               config.ConfigStore,
		findNewestBuild:           config.FindNewestBuild,
		listPublishedBuilds:       config.ListPublishedBuilds,
		createBuild:               config.CreateBuild,
		deleteBuilds:              config.DeleteBuilds,
		getRepositoryContext:      config.GetRepositoryContext,
		findPrebuildWebhook:       config.FindPrebuildWebhook,
		unregisterPrebuildWebhook: config.UnregisterPrebuildWebhook,
		registerPrebuildWebhook:   config.RegisterPrebuildWebhook,
		getCommitsRange:           config.GetCommitsRange,
	}
}

func (s *WorkspaceConfigService) List(filter *stores.WorkspaceConfigFilter) ([]*models.WorkspaceConfig, error) {
	return s.configStore.List(filter)
}

func (s *WorkspaceConfigService) SetDefault(workspaceConfigName string) error {
	workspaceConfig, err := s.Find(&stores.WorkspaceConfigFilter{
		Name: &workspaceConfigName,
	})
	if err != nil {
		return err
	}

	defaultWorkspaceConfig, err := s.Find(&stores.WorkspaceConfigFilter{
		Url:     &workspaceConfig.RepositoryUrl,
		Default: util.Pointer(true),
	})
	if err != nil && !stores.IsWorkspaceConfigNotFound(err) {
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

func (s *WorkspaceConfigService) Find(filter *stores.WorkspaceConfigFilter) (*models.WorkspaceConfig, error) {
	if filter != nil && filter.Url != nil {
		cleanedUrl := util.CleanUpRepositoryUrl(*filter.Url)
		if !strings.HasSuffix(cleanedUrl, ".git") {
			cleanedUrl = cleanedUrl + ".git"
		}
		filter.Url = util.Pointer(cleanedUrl)
	}
	return s.configStore.Find(filter)
}

func (s *WorkspaceConfigService) Save(workspaceConfig *models.WorkspaceConfig) error {
	workspaceConfig.RepositoryUrl = util.CleanUpRepositoryUrl(workspaceConfig.RepositoryUrl)

	err := s.configStore.Save(workspaceConfig)
	if err != nil {
		return err
	}

	return s.SetDefault(workspaceConfig.Name)
}

func (s *WorkspaceConfigService) Delete(workspaceConfigName string, force bool) []error {
	wc, err := s.Find(&stores.WorkspaceConfigFilter{
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
