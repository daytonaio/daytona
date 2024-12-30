// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"testing"

	t_build "github.com/daytonaio/daytona/internal/testing/build"
	git_mocks "github.com/daytonaio/daytona/internal/testing/git/mocks"
	builder_mocks "github.com/daytonaio/daytona/internal/testing/server/workspaces/mocks"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/stretchr/testify/suite"
)

var expectedBuilds []*build.Build

type BuilderTestSuite struct {
	suite.Suite
	mockGitService *git_mocks.MockGitService
	mockBuildStore build.Store
	builder        build.IBuilder
}

func NewBuilderTestSuite() *BuilderTestSuite {
	return &BuilderTestSuite{}
}

func (s *BuilderTestSuite) SetupTest() {
	s.mockBuildStore = t_build.NewInMemoryBuildStore()
	s.mockGitService = git_mocks.NewMockGitService()
	factory := build.NewBuilderFactory(build.BuilderFactoryConfig{
		BuildStore: s.mockBuildStore,
	})
	s.builder, _ = factory.Create(*builder_mocks.MockBuild, "")
	err := s.mockBuildStore.Save(builder_mocks.MockBuild)
	if err != nil {
		panic(err)
	}
}

func TestBuilder(t *testing.T) {
	suite.Run(t, NewBuilderTestSuite())
}

func (s *BuilderTestSuite) TestSaveBuild() {
	expectedBuilds = append(expectedBuilds, builder_mocks.MockBuild)

	require := s.Require()

	err := s.mockBuildStore.Save(builder_mocks.MockBuild)
	require.NoError(err)

	savedBuilds, err := s.mockBuildStore.List(nil)
	require.NoError(err)
	require.ElementsMatch(expectedBuilds, savedBuilds)
}
