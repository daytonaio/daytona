// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfigs_test

import (
	"context"
	"testing"

	git_provider_mock "github.com/daytonaio/daytona/internal/testing/gitprovider/mocks"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	workspaceconfig_internal "github.com/daytonaio/daytona/internal/testing/server/workspaceconfig"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfigs"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/stretchr/testify/suite"
)

var workspaceConfig1Image = "image1"
var workspaceConfig1User = "user1"

var workspaceConfig1 = &models.WorkspaceConfig{
	Name:          "wc1",
	Image:         workspaceConfig1Image,
	User:          workspaceConfig1User,
	BuildConfig:   nil,
	RepositoryUrl: repository1.Url,
	IsDefault:     true,
	Prebuilds: []*models.PrebuildConfig{
		prebuild1,
		prebuild2,
	},
}

var workspaceConfig2 = &models.WorkspaceConfig{
	Name:          "wc2",
	Image:         "image2",
	User:          "user2",
	BuildConfig:   nil,
	RepositoryUrl: "https://github.com/daytonaio/daytona.git",
}

var workspaceConfig3 = &models.WorkspaceConfig{
	Name:          "wc3",
	Image:         "image3",
	User:          "user3",
	BuildConfig:   nil,
	RepositoryUrl: "https://github.com/daytonaio/daytona3.git",
}

var workspaceConfig4 = &models.WorkspaceConfig{
	Name:          "wc4",
	Image:         "image4",
	User:          "user4",
	BuildConfig:   nil,
	RepositoryUrl: "https://github.com/daytonaio/daytona4.git",
}

var expectedWorkspaceConfigs []*models.WorkspaceConfig
var expectedFilteredWorkspaceConfigs []*models.WorkspaceConfig

var expectedWorkspaceConfigsMap map[string]*models.WorkspaceConfig
var expectedFilteredWorkspaceConfigsMap map[string]*models.WorkspaceConfig

type WorkspaceConfigServiceTestSuite struct {
	suite.Suite
	workspaceConfigService services.IWorkspaceConfigService
	workspaceConfigStore   stores.WorkspaceConfigStore
	gitProviderService     mocks.MockGitProviderService
	buildService           mocks.MockBuildService
	gitProvider            git_provider_mock.MockGitProvider
}

func NewConfigServiceTestSuite() *WorkspaceConfigServiceTestSuite {
	return &WorkspaceConfigServiceTestSuite{}
}

func (s *WorkspaceConfigServiceTestSuite) SetupTest() {
	expectedWorkspaceConfigs = []*models.WorkspaceConfig{
		workspaceConfig1, workspaceConfig2, workspaceConfig3,
	}

	expectedPrebuilds = []*models.PrebuildConfig{
		prebuild1, prebuild2,
	}

	expectedWorkspaceConfigsMap = map[string]*models.WorkspaceConfig{
		workspaceConfig1.Name: workspaceConfig1,
		workspaceConfig2.Name: workspaceConfig2,
		workspaceConfig3.Name: workspaceConfig3,
	}

	expectedPrebuildsMap = map[string]*models.PrebuildConfig{
		prebuild1.Id: prebuild1,
		prebuild2.Id: prebuild2,
	}

	expectedFilteredWorkspaceConfigs = []*models.WorkspaceConfig{
		workspaceConfig1, workspaceConfig2,
	}

	expectedFilteredPrebuilds = []*models.PrebuildConfig{
		prebuild1,
	}

	expectedFilteredWorkspaceConfigsMap = map[string]*models.WorkspaceConfig{
		workspaceConfig1.Name: workspaceConfig1,
		workspaceConfig2.Name: workspaceConfig2,
	}

	expectedFilteredPrebuildsMap = map[string]*models.PrebuildConfig{
		prebuild1.Id: prebuild1,
	}

	s.workspaceConfigStore = workspaceconfig_internal.NewInMemoryWorkspaceConfigStore()
	s.workspaceConfigService = workspaceconfigs.NewWorkspaceConfigService(workspaceconfigs.WorkspaceConfigServiceConfig{
		ConfigStore: s.workspaceConfigStore,
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
		CreateBuild: func(ctx context.Context, workspaceConfig *models.WorkspaceConfig, repo *gitprovider.GitRepository, prebuildId string) error {
			createBuildDto := services.CreateBuildDTO{
				WorkspaceConfigName: workspaceConfig.Name,
				Branch:              repo.Branch,
				PrebuildId:          &prebuildId,
				EnvVars:             workspaceConfig.EnvVars,
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

	for _, wc := range expectedWorkspaceConfigs {
		_ = s.workspaceConfigStore.Save(wc)
	}
}

func TestWorkspaceConfigService(t *testing.T) {
	suite.Run(t, NewConfigServiceTestSuite())
}

func (s *WorkspaceConfigServiceTestSuite) TestList() {
	require := s.Require()

	workspaceConfigs, err := s.workspaceConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedWorkspaceConfigs, workspaceConfigs)
}

func (s *WorkspaceConfigServiceTestSuite) TestFind() {
	require := s.Require()

	workspaceConfig, err := s.workspaceConfigService.Find(&stores.WorkspaceConfigFilter{
		Name: &workspaceConfig1.Name,
	})
	require.Nil(err)
	require.Equal(workspaceConfig1, workspaceConfig)
}
func (s *WorkspaceConfigServiceTestSuite) TestSetDefault() {
	require := s.Require()

	err := s.workspaceConfigService.SetDefault(workspaceConfig2.Name)
	require.Nil(err)

	workspaceConfig, err := s.workspaceConfigService.Find(&stores.WorkspaceConfigFilter{
		Url:     util.Pointer(workspaceConfig1.RepositoryUrl),
		Default: util.Pointer(true),
	})
	require.Nil(err)

	require.Equal(workspaceConfig2, workspaceConfig)
}

func (s *WorkspaceConfigServiceTestSuite) TestSave() {
	expectedWorkspaceConfigs = append(expectedWorkspaceConfigs, workspaceConfig4)

	require := s.Require()

	err := s.workspaceConfigService.Save(workspaceConfig4)
	require.Nil(err)

	workspaceConfigs, err := s.workspaceConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedWorkspaceConfigs, workspaceConfigs)
}

func (s *WorkspaceConfigServiceTestSuite) TestDelete() {
	expectedWorkspaceConfigs = expectedWorkspaceConfigs[:2]

	require := s.Require()

	err := s.workspaceConfigService.Delete(workspaceConfig3.Name, false)
	require.Nil(err)

	workspaceConfigs, errs := s.workspaceConfigService.List(nil)
	require.Nil(errs)
	require.ElementsMatch(expectedWorkspaceConfigs, workspaceConfigs)
}

func (s *WorkspaceConfigServiceTestSuite) AfterTest(_, _ string) {
	s.gitProviderService.AssertExpectations(s.T())
	s.gitProviderService.ExpectedCalls = nil
	s.buildService.AssertExpectations(s.T())
	s.buildService.ExpectedCalls = nil
	s.gitProvider.AssertExpectations(s.T())
	s.gitProvider.ExpectedCalls = nil
}
