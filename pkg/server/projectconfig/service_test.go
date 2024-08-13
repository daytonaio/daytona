// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig_test

import (
	"testing"

	projectconfig_internal "github.com/daytonaio/daytona/internal/testing/server/projectconfig"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/projectconfig"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	"github.com/stretchr/testify/suite"
)

var projectConfig1Image = "image1"
var projectConfig1User = "user1"

var projectConfig1 *config.ProjectConfig = &config.ProjectConfig{
	Name:        "pc1",
	Image:       projectConfig1Image,
	User:        projectConfig1User,
	BuildConfig: nil,
	Repository: &gitprovider.GitRepository{
		Url: "url1",
	},
	IsDefault: true,
}

var projectConfig2 *config.ProjectConfig = &config.ProjectConfig{
	Name:        "pc2",
	Image:       "image2",
	User:        "user2",
	BuildConfig: nil,
	Repository: &gitprovider.GitRepository{
		Url: "url1",
	},
}

var projectConfig3 *config.ProjectConfig = &config.ProjectConfig{
	Name:        "pc3",
	Image:       "image3",
	User:        "user3",
	BuildConfig: nil,
	Repository: &gitprovider.GitRepository{
		Url: "url3",
	},
}

var projectConfig4 *config.ProjectConfig = &config.ProjectConfig{
	Name:        "pc4",
	Image:       "image4",
	User:        "user4",
	BuildConfig: nil,
	Repository: &gitprovider.GitRepository{
		Url: "url4",
	},
}

var expectedProjectConfigs []*config.ProjectConfig
var expectedFilteredProjectConfigs []*config.ProjectConfig

var expectedProjectConfigsMap map[string]*config.ProjectConfig
var expectedFilteredProjectConfigsMap map[string]*config.ProjectConfig

type ProjectConfigServiceTestSuite struct {
	suite.Suite
	projectConfigService projectconfig.IProjectConfigService
	projectConfigStore   config.Store
}

func NewConfigServiceTestSuite() *ProjectConfigServiceTestSuite {
	return &ProjectConfigServiceTestSuite{}
}

func (s *ProjectConfigServiceTestSuite) SetupTest() {
	expectedProjectConfigs = []*config.ProjectConfig{
		projectConfig1, projectConfig2, projectConfig3,
	}

	expectedProjectConfigsMap = map[string]*config.ProjectConfig{
		projectConfig1.Name: projectConfig1,
		projectConfig2.Name: projectConfig2,
		projectConfig3.Name: projectConfig3,
	}

	expectedFilteredProjectConfigs = []*config.ProjectConfig{
		projectConfig1, projectConfig2,
	}

	expectedFilteredProjectConfigsMap = map[string]*config.ProjectConfig{
		projectConfig1.Name: projectConfig1,
		projectConfig2.Name: projectConfig2,
	}

	s.projectConfigStore = projectconfig_internal.NewInMemoryProjectConfigStore()
	s.projectConfigService = projectconfig.NewConfigService(projectconfig.ProjectConfigServiceConfig{
		ConfigStore: s.projectConfigStore,
	})

	for _, pc := range expectedProjectConfigs {
		_ = s.projectConfigStore.Save(pc)
	}
}

func TestProjectConfigService(t *testing.T) {
	suite.Run(t, NewConfigServiceTestSuite())
}

func (s *ProjectConfigServiceTestSuite) TestList() {
	require := s.Require()

	projectConfigs, err := s.projectConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedProjectConfigs, projectConfigs)
}

func (s *ProjectConfigServiceTestSuite) TestFind() {
	require := s.Require()

	projectConfig, err := s.projectConfigService.Find(&config.Filter{
		Name: &projectConfig1.Name,
	})
	require.Nil(err)
	require.Equal(projectConfig1, projectConfig)
}
func (s *ProjectConfigServiceTestSuite) TestSetDefault() {
	require := s.Require()

	err := s.projectConfigService.SetDefault(projectConfig2.Name)
	require.Nil(err)

	projectConfig, err := s.projectConfigService.Find(&config.Filter{
		Url:     util.Pointer("url1"),
		Default: util.Pointer(true),
	})
	require.Nil(err)

	require.Equal(projectConfig2, projectConfig)
}

func (s *ProjectConfigServiceTestSuite) TestSave() {
	expectedProjectConfigs = append(expectedProjectConfigs, projectConfig4)

	require := s.Require()

	err := s.projectConfigService.Save(projectConfig4)
	require.Nil(err)

	projectConfigs, err := s.projectConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedProjectConfigs, projectConfigs)
}

func (s *ProjectConfigServiceTestSuite) TestDelete() {
	expectedProjectConfigs = expectedProjectConfigs[:2]

	require := s.Require()

	err := s.projectConfigService.Delete(projectConfig3.Name)
	require.Nil(err)

	projectConfigs, err := s.projectConfigService.List(nil)
	require.Nil(err)
	require.ElementsMatch(expectedProjectConfigs, projectConfigs)
}
