// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"testing"

	"github.com/daytonaio/daytona/internal/testing/git/mocks"
	t_build "github.com/daytonaio/daytona/internal/testing/server/build"
	build_mocks "github.com/daytonaio/daytona/internal/testing/server/workspaces/mocks"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/stretchr/testify/suite"
)

var expectedBuilds []*build.Build

type BuilderTestSuite struct {
	suite.Suite
	mockGitService *mocks.MockGitService
	builder        build.IBuilder
	buildStore     build.Store
}

func NewBuilderTestSuite() *BuilderTestSuite {
	return &BuilderTestSuite{}
}

func (s *BuilderTestSuite) SetupTest() {
	s.buildStore = t_build.NewInMemoryBuildStore()
	s.mockGitService = mocks.NewMockGitService()
	factory := build.NewBuilderFactory(build.BuilderFactoryConfig{
		BuildStore: s.buildStore,
	})
	s.builder, _ = factory.Create(*build_mocks.MockBuild)
	err := s.buildStore.Save(build_mocks.MockBuild)
	if err != nil {
		panic(err)
	}
}

func TestBuilder(t *testing.T) {
	suite.Run(t, NewBuilderTestSuite())
}

func (s *BuilderTestSuite) TestSaveBuild() {
	expectedBuilds = append(expectedBuilds, build_mocks.MockBuild)

	require := s.Require()

	err := s.builder.SaveBuild(*build_mocks.MockBuild)
	require.NoError(err)

	savedBuilds, err := s.buildStore.List(nil)
	require.NoError(err)
	require.ElementsMatch(expectedBuilds, savedBuilds)
}
