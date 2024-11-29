// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplates_test

import (
	"context"
	"testing"

	git_provider_mock "github.com/daytonaio/daytona/internal/testing/gitprovider/mocks"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	workspacetemplate_internal "github.com/daytonaio/daytona/internal/testing/server/workspacetemplate"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/workspacetemplates"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/stretchr/testify/suite"
)

var workspaceTemplate1Image = "image1"
var workspaceTemplate1User = "user1"

var workspaceTemplate1 = &models.WorkspaceTemplate{
	Name:          "wt1",
	Image:         workspaceTemplate1Image,
	User:          workspaceTemplate1User,
	BuildConfig:   nil,
	RepositoryUrl: repository1.Url,
	IsDefault:     true,
	Prebuilds: []*models.PrebuildConfig{
		prebuild1,
		prebuild2,
	},
}

var workspaceTemplate2 = &models.WorkspaceTemplate{
	Name:          "wt2",
	Image:         "image2",
	User:          "user2",
	BuildConfig:   nil,
	RepositoryUrl: "https://github.com/daytonaio/daytona.git",
}

var workspaceTemplate3 = &models.WorkspaceTemplate{
	Name:          "wt3",
	Image:         "image3",
	User:          "user3",
	BuildConfig:   nil,
	RepositoryUrl: "https://github.com/daytonaio/daytona3.git",
}

var workspaceTemplate4 = &models.WorkspaceTemplate{
	Name:          "wt4",
	Image:         "image4",
	User:          "user4",
	BuildConfig:   nil,
	RepositoryUrl: "https://github.com/daytonaio/daytona4.git",
}

var expectedWorkspaceTemplates []*models.WorkspaceTemplate
var expectedFilteredWorkspaceTemplates []*models.WorkspaceTemplate

var expectedWorkspaceTemplatesMap map[string]*models.WorkspaceTemplate
var expectedFilteredWorkspaceTemplatesMap map[string]*models.WorkspaceTemplate

type WorkspaceTemplateServiceTestSuite struct {
	suite.Suite
	workspaceTemplateService services.IWorkspaceTemplateService
	workspaceTemplateStore   stores.WorkspaceTemplateStore
	gitProviderService       mocks.MockGitProviderService
	buildService             mocks.MockBuildService
	gitProvider              git_provider_mock.MockGitProvider
}

func NewConfigServiceTestSuite() *WorkspaceTemplateServiceTestSuite {
	return &WorkspaceTemplateServiceTestSuite{}
}

func (s *WorkspaceTemplateServiceTestSuite) SetupTest() {
	expectedWorkspaceTemplates = []*models.WorkspaceTemplate{
		workspaceTemplate1, workspaceTemplate2, workspaceTemplate3,
	}

	expectedPrebuilds = []*models.PrebuildConfig{
		prebuild1, prebuild2,
	}

	expectedWorkspaceTemplatesMap = map[string]*models.WorkspaceTemplate{
		workspaceTemplate1.Name: workspaceTemplate1,
		workspaceTemplate2.Name: workspaceTemplate2,
		workspaceTemplate3.Name: workspaceTemplate3,
	}

	expectedPrebuildsMap = map[string]*models.PrebuildConfig{
		prebuild1.Id: prebuild1,
		prebuild2.Id: prebuild2,
	}

	expectedFilteredWorkspaceTemplates = []*models.WorkspaceTemplate{
		workspaceTemplate1, workspaceTemplate2,
	}

	expectedFilteredPrebuilds = []*models.PrebuildConfig{
		prebuild1,
	}

	expectedFilteredWorkspaceTemplatesMap = map[string]*models.WorkspaceTemplate{
		workspaceTemplate1.Name: workspaceTemplate1,
		workspaceTemplate2.Name: workspaceTemplate2,
	}

	expectedFilteredPrebuildsMap = map[string]*models.PrebuildConfig{
		prebuild1.Id: prebuild1,
	}

	s.workspaceTemplateStore = workspacetemplate_internal.NewInMemoryWorkspaceTemplateStore()
	s.workspaceTemplateService = workspacetemplates.NewWorkspaceTemplateService(workspacetemplates.WorkspaceTemplateServiceConfig{
		ConfigStore: s.workspaceTemplateStore,
		FindNewestBuild: func(ctx context.Context, prebuildId string) (*models.Build, error) {
			return s.buildService.Find(&stores.BuildFilter{
				PrebuildIds: &[]string{prebuildId},
				GetNewest:   util.Pointer(true),
			})
		},
		ListPublishedBuilds: func(ctx context.Context) ([]*models.Build, error) {
			return s.buildService.List(&stores.BuildFilter{
				States: &[]models.BuildState{models.BuildStatePublished},
			})
		},
		CreateBuild: func(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate, repo *gitprovider.GitRepository, prebuildId string) error {
			createBuildDto := services.CreateBuildDTO{
				WorkspaceTemplateName: workspaceTemplate.Name,
				Branch:                repo.Branch,
				PrebuildId:            &prebuildId,
				EnvVars:               workspaceTemplate.EnvVars,
			}

			_, err := s.buildService.Create(createBuildDto)
			return err
		},
		DeleteBuilds: func(ctx context.Context, id, prebuildId *string, force bool) []error {
			var prebuildIds *[]string
			if prebuildId != nil {
				prebuildIds = &[]string{*prebuildId}
			}

			return s.buildService.MarkForDeletion(&stores.BuildFilter{
				Id:          id,
				PrebuildIds: prebuildIds,
			}, force)
		},
		GetRepositoryContext: func(ctx context.Context, url string) (repo *gitprovider.GitRepository, gitProviderId string, err error) {
			gitProvider, gitProviderId, err := s.gitProviderService.GetGitProviderForUrl(url)
			if err != nil {
				return nil, "", err
			}

			repo, err = gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
				Url: url,
			})

			return repo, gitProviderId, err
		},
		FindPrebuildWebhook: func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error) {
			return s.gitProviderService.GetPrebuildWebhook(gitProviderId, repo, endpointUrl)
		},
		UnregisterPrebuildWebhook: func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, id string) error {
			return s.gitProviderService.UnregisterPrebuildWebhook(gitProviderId, repo, id)
		},
		RegisterPrebuildWebhook: func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error) {
			return s.gitProviderService.RegisterPrebuildWebhook(gitProviderId, repo, endpointUrl)
		},
		GetCommitsRange: func(ctx context.Context, repo *gitprovider.GitRepository, initialSha string, currentSha string) (int, error) {
			gitProvider, _, err := s.gitProviderService.GetGitProviderForUrl(repo.Url)
			if err != nil {
				return 0, err
			}

			return gitProvider.GetCommitsRange(repo, initialSha, currentSha)
		},
	})

	for _, wt := range expectedWorkspaceTemplates {
		_ = s.workspaceTemplateStore.Save(wt)
	}
}

func TestWorkspaceTemplateService(t *testing.T) {
	suite.Run(t, NewConfigServiceTestSuite())
}

func (s *WorkspaceTemplateServiceTestSuite) TestList() {
	require := s.Require()

	workspaceTemplates, err := s.workspaceTemplateService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedWorkspaceTemplates, workspaceTemplates)
}

func (s *WorkspaceTemplateServiceTestSuite) TestFind() {
	require := s.Require()

	workspaceTemplate, err := s.workspaceTemplateService.Find(&stores.WorkspaceTemplateFilter{
		Name: &workspaceTemplate1.Name,
	})
	require.Nil(err)
	require.Equal(workspaceTemplate1, workspaceTemplate)
}
func (s *WorkspaceTemplateServiceTestSuite) TestSetDefault() {
	require := s.Require()

	err := s.workspaceTemplateService.SetDefault(workspaceTemplate2.Name)
	require.Nil(err)

	workspaceTemplate, err := s.workspaceTemplateService.Find(&stores.WorkspaceTemplateFilter{
		Url:     util.Pointer(workspaceTemplate1.RepositoryUrl),
		Default: util.Pointer(true),
	})
	require.Nil(err)

	require.Equal(workspaceTemplate2, workspaceTemplate)
}

func (s *WorkspaceTemplateServiceTestSuite) TestSave() {
	expectedWorkspaceTemplates = append(expectedWorkspaceTemplates, workspaceTemplate4)

	require := s.Require()

	err := s.workspaceTemplateService.Save(workspaceTemplate4)
	require.Nil(err)

	workspaceTemplates, err := s.workspaceTemplateService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedWorkspaceTemplates, workspaceTemplates)
}

func (s *WorkspaceTemplateServiceTestSuite) TestDelete() {
	expectedWorkspaceTemplates = expectedWorkspaceTemplates[:2]

	require := s.Require()

	err := s.workspaceTemplateService.Delete(workspaceTemplate3.Name, false)
	require.Nil(err)

	workspaceTemplates, errs := s.workspaceTemplateService.List(nil)
	require.Nil(errs)
	require.ElementsMatch(expectedWorkspaceTemplates, workspaceTemplates)
}

func (s *WorkspaceTemplateServiceTestSuite) AfterTest(_, _ string) {
	s.gitProviderService.AssertExpectations(s.T())
	s.gitProviderService.ExpectedCalls = nil
	s.buildService.AssertExpectations(s.T())
	s.buildService.ExpectedCalls = nil
	s.gitProvider.AssertExpectations(s.T())
	s.gitProvider.ExpectedCalls = nil
}
