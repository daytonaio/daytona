// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"testing"

	t_build "github.com/daytonaio/daytona/internal/testing/server/build"
	"github.com/daytonaio/daytona/internal/testing/server/workspaces/mocks"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/poller"
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
	Poller             poller.IPoller
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
	s.Poller = build.NewBuildRunner(build.BuildRunnerInstanceConfig{
		Scheduler:      &s.mockScheduler,
		Interval:       "0 */5 * * * *",
		BuilderFactory: &s.mockBuilderFactory,
		BuildStore:     s.buildStore,
		LoggerFactory:  s.loggerFactory,
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

	err := s.Poller.Start()
	require.NoError(err)

	s.mockScheduler.AssertExpectations(s.T())
}

func (s *BuildRunnerTestSuite) TestStop() {
	s.mockScheduler.On("Stop").Return()

	s.Poller.Stop()

	s.mockScheduler.AssertExpectations(s.T())
}

// TODO FIXME: Need to figure out how to test the runBuildProcess goroutine
func (s *BuildRunnerTestSuite) TestPoll() {
	s.T().Skip("Need to figure out how to test the runBuildProcess goroutine")

	s.mockBuilderFactory.On("Create", mocks.MockBuild).Return(&s.mockBuilder, nil)
	s.mockBuilder.On("Build", mocks.MockBuild).Return(mocks.MockBuild, nil)
	s.mockBuilder.On("Publish", mocks.MockBuild).Return(nil)
	s.mockBuilder.On("CleanUp").Return(nil)

	s.Poller.Poll()

	s.mockBuilderFactory.AssertExpectations(s.T())
	s.mockBuilder.AssertExpectations(s.T())
}
