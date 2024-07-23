// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"io"
	"testing"

	"github.com/daytonaio/daytona/internal/testing/git/mocks"
	t_build "github.com/daytonaio/daytona/internal/testing/server/build"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	project_build "github.com/daytonaio/daytona/pkg/workspace/project/build"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var p project.Project = project.Project{
	ProjectConfig: config.ProjectConfig{
		Repository: &gitprovider.GitRepository{},
		Build: &project_build.ProjectBuildConfig{
			Devcontainer: &project_build.DevcontainerConfig{
				FilePath: ".devcontainer/devcontainer.json",
			},
		}},
}

var predefBuildResult build.BuildResult = build.BuildResult{
	Hash:              "test-predef",
	User:              "test-predef",
	ImageName:         "test-predef",
	ProjectVolumePath: "test-predef",
}

var buildResult build.BuildResult = build.BuildResult{
	Hash:              "test",
	User:              "test",
	ImageName:         "test",
	ProjectVolumePath: "test",
}

var expectedResults []*build.BuildResult

type BuilderTestSuite struct {
	suite.Suite
	mockGitService   *mocks.MockGitService
	builder          build.IBuilder
	buildResultStore build.Store
}

func NewBuilderTestSuite() *BuilderTestSuite {
	return &BuilderTestSuite{}
}

func (s *BuilderTestSuite) SetupTest() {
	s.buildResultStore = t_build.NewInMemoryBuildStore()
	s.mockGitService = mocks.NewMockGitService()
	factory := build.NewBuilderFactory(build.BuilderFactoryConfig{
		BuilderConfig: build.BuilderConfig{
			BuildResultStore: s.buildResultStore,
		},
		CreateGitService: func(projectDir string, w io.Writer) git.IGitService {
			return s.mockGitService
		},
	})
	s.mockGitService.On("CloneRepository", mock.Anything, mock.Anything).Return(nil)
	s.builder, _ = factory.Create(p, nil)
	err := s.buildResultStore.Save(&predefBuildResult)
	if err != nil {
		panic(err)
	}
	expectedResults = append(expectedResults, &predefBuildResult)
}

func TestBuilder(t *testing.T) {
	suite.Run(t, NewBuilderTestSuite())
}

func (s *BuilderTestSuite) TestSaveBuildResults() {
	expectedResults := append(expectedResults, &buildResult)

	require := s.Require()

	err := s.builder.SaveBuildResults(buildResult)
	require.NoError(err)

	savedResults, err := s.buildResultStore.List()
	require.NoError(err)
	require.ElementsMatch(expectedResults, savedResults)
}
