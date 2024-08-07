// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"testing"

	t_build "github.com/daytonaio/daytona/internal/testing/server/build"
	"github.com/daytonaio/daytona/internal/testing/server/workspaces/mocks"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BuildRunnerTestSuite struct {
	suite.Suite
	mockBuilderFactory mocks.MockBuilderFactory
	mockBuilder        mocks.MockBuilder
	mockScheduler      mocks.MockScheduler
	loggerFactory      logs.LoggerFactory
	buildStore         build.Store
	Runner             build.BuildRunner
}

func NewBuildRunnerTestSuite() *BuildRunnerTestSuite {
	return &BuildRunnerTestSuite{}
}

func TestBuildRunner(t *testing.T) {
	s := NewBuildRunnerTestSuite()

	s.mockBuilderFactory = mocks.MockBuilderFactory{}
	s.mockBuilder = mocks.MockBuilder{}
	s.mockScheduler = mocks.MockScheduler{}

	s.buildStore = t_build.NewInMemoryBuildStore()
	s.loggerFactory = logs.NewLoggerFactory(t.TempDir())

	s.Runner = *build.NewBuildRunner(build.BuildRunnerInstanceConfig{
		Interval:         "0 */5 * * * *",
		Scheduler:        &s.mockScheduler,
		BuildRunnerId:    "1",
		BuildStore:       s.buildStore,
		BuilderFactory:   &s.mockBuilderFactory,
		LoggerFactory:    s.loggerFactory,
		TelemetryEnabled: false,
	})

	suite.Run(t, s)
}

func (s *BuildRunnerTestSuite) SetupTest() {
	err := s.buildStore.Save(mocks.MockBuild)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *BuildRunnerTestSuite) TestStart() {
	s.mockScheduler.On("AddFunc", mock.Anything, mock.Anything).Return(nil)
	s.mockScheduler.On("Start").Return()

	require := s.Require()

	err := s.Runner.Start()
	require.NoError(err)

	s.mockScheduler.AssertExpectations(s.T())
}

func (s *BuildRunnerTestSuite) TestStop() {
	s.mockScheduler.On("Stop").Return()

	s.Runner.Stop()

	s.mockScheduler.AssertExpectations(s.T())
}

// TODO FIXME: Need to figure out how to test the RunBuildProcess goroutine
func (s *BuildRunnerTestSuite) TestRun() {
	s.T().Skip("Need to figure out how to test the runBuildProcess goroutine")
}

func (s *BuildRunnerTestSuite) TestRunBuildProcess() {
	pendingBuild := *mocks.MockBuild
	s.mockBuilderFactory.On("Create", pendingBuild).Return(&s.mockBuilder, nil)

	runningBuild := *mocks.MockBuild
	runningBuild.State = build.BuildStateRunning
	s.mockBuilder.On("Build", runningBuild).Return("image", "user", nil)

	successBuild := *mocks.MockBuild
	successBuild.State = build.BuildStateSuccess
	successBuild.Image = "image"
	successBuild.User = "user"
	s.mockBuilder.On("Publish", successBuild).Return(nil)

	s.mockBuilder.On("CleanUp").Return(nil)

	s.Runner.RunBuildProcess(mocks.MockBuild, nil)

	s.mockBuilderFactory.AssertExpectations(s.T())
	s.mockBuilder.AssertExpectations(s.T())

	s.Require().Equal(mocks.MockBuild.Image, "image")
	s.Require().Equal(mocks.MockBuild.User, "user")
	s.Require().Equal(mocks.MockBuild.State, build.BuildStatePublished)
}
