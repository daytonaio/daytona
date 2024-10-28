// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig_test

import (
	"testing"

	git_provider_mock "github.com/daytonaio/daytona/internal/testing/gitprovider/mocks"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	workspaceconfig_internal "github.com/daytonaio/daytona/internal/testing/server/workspaceconfig"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfig"
	"github.com/daytonaio/daytona/pkg/target/workspace/config"
	"github.com/stretchr/testify/suite"
)

var workspaceConfig1Image = "image1"
var workspaceConfig1User = "user1"

var workspaceConfig1 *config.WorkspaceConfig = &config.WorkspaceConfig{
	Name:          "wc1",
	Image:         workspaceConfig1Image,
	User:          workspaceConfig1User,
	BuildConfig:   nil,
	RepositoryUrl: repository1.Url,
	IsDefault:     true,
	Prebuilds: []*config.PrebuildConfig{
		prebuild1,
		prebuild2,
	},
}

var workspaceConfig2 *config.WorkspaceConfig = &config.WorkspaceConfig{
	Name:          "wc2",
	Image:         "image2",
	User:          "user2",
	BuildConfig:   nil,
	RepositoryUrl: "https://github.com/daytonaio/daytona.git",
}

var workspaceConfig3 *config.WorkspaceConfig = &config.WorkspaceConfig{
	Name:          "wc3",
	Image:         "image3",
	User:          "user3",
	BuildConfig:   nil,
	RepositoryUrl: "https://github.com/daytonaio/daytona3.git",
}

var workspaceConfig4 *config.WorkspaceConfig = &config.WorkspaceConfig{
	Name:          "wc4",
	Image:         "image4",
	User:          "user4",
	BuildConfig:   nil,
	RepositoryUrl: "https://github.com/daytonaio/daytona4.git",
}

var expectedWorkspaceConfigs []*config.WorkspaceConfig
var expectedFilteredWorkspaceConfigs []*config.WorkspaceConfig

var expectedWorkspaceConfigsMap map[string]*config.WorkspaceConfig
var expectedFilteredWorkspaceConfigsMap map[string]*config.WorkspaceConfig

type WorkspaceConfigServiceTestSuite struct {
	suite.Suite
	workspaceConfigService workspaceconfig.IWorkspaceConfigService
	workspaceConfigStore   config.Store
	gitProviderService     mocks.MockGitProviderService
	buildService           mocks.MockBuildService
	gitProvider            git_provider_mock.MockGitProvider
}

func NewConfigServiceTestSuite() *WorkspaceConfigServiceTestSuite {
	return &WorkspaceConfigServiceTestSuite{}
}

func (s *WorkspaceConfigServiceTestSuite) SetupTest() {
	expectedWorkspaceConfigs = []*config.WorkspaceConfig{
		workspaceConfig1, workspaceConfig2, workspaceConfig3,
	}

	expectedPrebuilds = []*config.PrebuildConfig{
		prebuild1, prebuild2,
	}

	expectedWorkspaceConfigsMap = map[string]*config.WorkspaceConfig{
		workspaceConfig1.Name: workspaceConfig1,
		workspaceConfig2.Name: workspaceConfig2,
		workspaceConfig3.Name: workspaceConfig3,
	}

	expectedPrebuildsMap = map[string]*config.PrebuildConfig{
		prebuild1.Id: prebuild1,
		prebuild2.Id: prebuild2,
	}

	expectedFilteredWorkspaceConfigs = []*config.WorkspaceConfig{
		workspaceConfig1, workspaceConfig2,
	}

	expectedFilteredPrebuilds = []*config.PrebuildConfig{
		prebuild1,
	}

	expectedFilteredWorkspaceConfigsMap = map[string]*config.WorkspaceConfig{
		workspaceConfig1.Name: workspaceConfig1,
		workspaceConfig2.Name: workspaceConfig2,
	}

	expectedFilteredPrebuildsMap = map[string]*config.PrebuildConfig{
		prebuild1.Id: prebuild1,
	}

	s.workspaceConfigStore = workspaceconfig_internal.NewInMemoryWorkspaceConfigStore()
	s.workspaceConfigService = workspaceconfig.NewWorkspaceConfigService(workspaceconfig.WorkspaceConfigServiceConfig{
		ConfigStore:        s.workspaceConfigStore,
		GitProviderService: &s.gitProviderService,
		BuildService:       &s.buildService,
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

	workspaceConfig, err := s.workspaceConfigService.Find(&config.WorkspaceConfigFilter{
		Name: &workspaceConfig1.Name,
	})
	require.Nil(err)
	require.Equal(workspaceConfig1, workspaceConfig)
}
func (s *WorkspaceConfigServiceTestSuite) TestSetDefault() {
	require := s.Require()

	err := s.workspaceConfigService.SetDefault(workspaceConfig2.Name)
	require.Nil(err)

	workspaceConfig, err := s.workspaceConfigService.Find(&config.WorkspaceConfigFilter{
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
