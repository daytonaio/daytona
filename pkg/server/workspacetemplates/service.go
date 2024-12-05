// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplates

import (
	"context"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

type WorkspaceTemplateServiceConfig struct {
	PrebuildWebhookEndpoint string
	ConfigStore             stores.WorkspaceTemplateStore

	FindNewestBuild           func(ctx context.Context, prebuildId string) (*services.BuildDTO, error)
	ListSuccessfulBuilds      func(ctx context.Context) ([]*services.BuildDTO, error)
	CreateBuild               func(ctx context.Context, wt *models.WorkspaceTemplate, repo *gitprovider.GitRepository, prebuildId string) error
	DeleteBuilds              func(ctx context.Context, id, prebuildId *string, force bool) []error
	GetRepositoryContext      func(ctx context.Context, url string) (repo *gitprovider.GitRepository, gitProviderId string, err error)
	FindPrebuildWebhook       func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error)
	UnregisterPrebuildWebhook func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, id string) error
	RegisterPrebuildWebhook   func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error)
	GetCommitsRange           func(ctx context.Context, repo *gitprovider.GitRepository, initialSha string, currentSha string) (int, error)
}

type WorkspaceTemplateService struct {
	prebuildWebhookEndpoint string
	templateStore           stores.WorkspaceTemplateStore

	findNewestBuild           func(ctx context.Context, prebuildId string) (*services.BuildDTO, error)
	listSuccessfulBuilds      func(ctx context.Context) ([]*services.BuildDTO, error)
	createBuild               func(ctx context.Context, wt *models.WorkspaceTemplate, repo *gitprovider.GitRepository, prebuildId string) error
	deleteBuilds              func(ctx context.Context, id, prebuildId *string, force bool) []error
	getRepositoryContext      func(ctx context.Context, url string) (repo *gitprovider.GitRepository, gitProviderId string, err error)
	getCommitsRange           func(ctx context.Context, repo *gitprovider.GitRepository, initialSha string, currentSha string) (int, error)
	findPrebuildWebhook       func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error)
	unregisterPrebuildWebhook func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, id string) error
	registerPrebuildWebhook   func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error)
}

func NewWorkspaceTemplateService(config WorkspaceTemplateServiceConfig) services.IWorkspaceTemplateService {
	return &WorkspaceTemplateService{
		prebuildWebhookEndpoint:   config.PrebuildWebhookEndpoint,
		templateStore:             config.ConfigStore,
		findNewestBuild:           config.FindNewestBuild,
		listSuccessfulBuilds:      config.ListSuccessfulBuilds,
		createBuild:               config.CreateBuild,
		deleteBuilds:              config.DeleteBuilds,
		getRepositoryContext:      config.GetRepositoryContext,
		findPrebuildWebhook:       config.FindPrebuildWebhook,
		unregisterPrebuildWebhook: config.UnregisterPrebuildWebhook,
		registerPrebuildWebhook:   config.RegisterPrebuildWebhook,
		getCommitsRange:           config.GetCommitsRange,
	}
}

func (s *WorkspaceTemplateService) List(ctx context.Context, filter *stores.WorkspaceTemplateFilter) ([]*models.WorkspaceTemplate, error) {
	return s.templateStore.List(ctx, filter)
}

func (s *WorkspaceTemplateService) SetDefault(ctx context.Context, workspaceTemplateName string) error {
	var err error
	ctx, err = s.templateStore.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer stores.RecoverAndRollback(ctx, s.templateStore)

	workspaceTemplate, err := s.Find(ctx, &stores.WorkspaceTemplateFilter{
		Name: &workspaceTemplateName,
	})
	if err != nil {
		return s.templateStore.RollbackTransaction(ctx, err)
	}

	defaultWorkspaceTemplate, err := s.Find(ctx, &stores.WorkspaceTemplateFilter{
		Url:     &workspaceTemplate.RepositoryUrl,
		Default: util.Pointer(true),
	})
	if err != nil && !stores.IsWorkspaceTemplateNotFound(err) {
		return s.templateStore.RollbackTransaction(ctx, err)
	}

	if defaultWorkspaceTemplate != nil {
		defaultWorkspaceTemplate.IsDefault = false
		err := s.templateStore.Save(ctx, defaultWorkspaceTemplate)
		if err != nil {
			return s.templateStore.RollbackTransaction(ctx, err)
		}
	}

	workspaceTemplate.IsDefault = true
	err = s.templateStore.Save(ctx, workspaceTemplate)
	if err != nil {
		return s.templateStore.RollbackTransaction(ctx, err)
	}

	return s.templateStore.CommitTransaction(ctx)
}

func (s *WorkspaceTemplateService) Find(ctx context.Context, filter *stores.WorkspaceTemplateFilter) (*models.WorkspaceTemplate, error) {
	if filter != nil && filter.Url != nil {
		cleanedUrl := util.CleanUpRepositoryUrl(*filter.Url)
		if !strings.HasSuffix(cleanedUrl, ".git") {
			cleanedUrl = cleanedUrl + ".git"
		}
		filter.Url = util.Pointer(cleanedUrl)
	}
	return s.templateStore.Find(ctx, filter)
}

func (s *WorkspaceTemplateService) Save(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate) error {
	workspaceTemplate.RepositoryUrl = util.CleanUpRepositoryUrl(workspaceTemplate.RepositoryUrl)

	err := s.templateStore.Save(ctx, workspaceTemplate)
	if err != nil {
		return err
	}

	return s.SetDefault(ctx, workspaceTemplate.Name)
}

func (s *WorkspaceTemplateService) Delete(ctx context.Context, workspaceTemplateName string, force bool) []error {
	wt, err := s.Find(ctx, &stores.WorkspaceTemplateFilter{
		Name: &workspaceTemplateName,
	})
	if err != nil {
		return []error{err}
	}

	// DeletePrebuild handles deleting the builds and removing the webhook
	for _, prebuild := range wt.Prebuilds {
		errs := s.DeletePrebuild(ctx, wt.Name, prebuild.Id, force)
		if len(errs) > 0 {
			return errs
		}
	}

	err = s.templateStore.Delete(ctx, wt)
	if err != nil {
		return []error{err}
	}

	return nil
}
