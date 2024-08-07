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
	mockBuilder        mocks.MockBuilderPlugin
	mockScheduler      mocks.MockSchedulerPlugin
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
	s.mockBuilder = mocks.MockBuilderPlugin{}
	s.mockScheduler = mocks.MockSchedulerPlugin{}

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

	s.mockBuilderFactory.On("Create", mocks.MockBuild).Return(&s.mockBuilder, nil)
	s.mockBuilder.On("Build", mocks.MockBuild).Return(mocks.MockBuild, nil)
	s.mockBuilder.On("Publish", mocks.MockBuild).Return(nil)
	s.mockBuilder.On("CleanUp").Return(nil)

	s.Runner.Run()

	s.mockBuilderFactory.AssertExpectations(s.T())
	s.mockBuilder.AssertExpectations(s.T())
}

// TODO FIXME: Need to figure out how to test the RunBuildProcess
func (s *BuildRunnerTestSuite) TestRunBuildProcess() {
	s.T().Skip("Need to figure out how to test the RunBuildProcess")

	s.mockBuilderFactory.On("Create", mocks.MockBuild).Return(&s.mockBuilder, nil)
	s.mockBuilder.On("Build", mocks.MockBuild).Return("image", "user", nil)
	s.mockBuilder.On("Publish", mocks.MockBuild).Return(nil)
	s.mockBuilder.On("CleanUp").Return(nil)

	s.Runner.RunBuildProcess(mocks.MockBuild, nil)

	s.mockBuilderFactory.AssertExpectations(s.T())
	s.mockBuilder.AssertExpectations(s.T())

	s.Require().Equal(mocks.MockBuild.Image, "image")
	s.Require().Equal(mocks.MockBuild.User, "user")
}
