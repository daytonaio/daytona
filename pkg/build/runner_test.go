// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"testing"

	t_build "github.com/daytonaio/daytona/internal/testing/build"
	git_mocks "github.com/daytonaio/daytona/internal/testing/git/mocks"
	logger_mocks "github.com/daytonaio/daytona/internal/testing/logger/mocks"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build"
	t_gitprovider "github.com/daytonaio/daytona/pkg/build/mocks"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var gitProviderConfig = gitprovider.GitProviderConfig{
	Id:         "github",
	Username:   "daytonaio",
	Token:      "",
	BaseApiUrl: nil,
}

type BuildRunnerTestSuite struct {
	suite.Suite
	mockBuilderFactory         mocks.MockBuilderFactory
	mockBuilder                mocks.MockBuilder
	mockScheduler              mocks.MockScheduler
	mockGitService             git_mocks.MockGitService
	loggerFactory              logs.LoggerFactory
	mockBuildStore             build.Store
	mockGitProviderConfigStore t_gitprovider.MockGitProviderConfigStore
	Runner                     build.BuildRunner
}

func NewBuildRunnerTestSuite() *BuildRunnerTestSuite {
	return &BuildRunnerTestSuite{}
}

func TestBuildRunner(t *testing.T) {
	s := NewBuildRunnerTestSuite()

	s.mockBuilderFactory = mocks.MockBuilderFactory{}
	s.mockBuilder = mocks.MockBuilder{}
	s.mockScheduler = mocks.MockScheduler{}
	s.mockGitProviderConfigStore = t_gitprovider.MockGitProviderConfigStore{}

	s.mockBuildStore = t_build.NewInMemoryBuildStore()
	logTempDir := t.TempDir()
	s.loggerFactory = logs.NewLoggerFactory(nil, &logTempDir)

	s.Runner = *build.NewBuildRunner(build.BuildRunnerInstanceConfig{
		Interval:         "0 */5 * * * *",
		Scheduler:        &s.mockScheduler,
		BuildRunnerId:    "1",
		BuildStore:       s.mockBuildStore,
		GitProviderStore: &s.mockGitProviderConfigStore,
		BuilderFactory:   &s.mockBuilderFactory,
		LoggerFactory:    s.loggerFactory,
		TelemetryEnabled: false,
	})

	suite.Run(t, s)
}

func (s *BuildRunnerTestSuite) SetupTest() {
	err := s.mockBuildStore.Save(mocks.MockBuild)
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
	s.mockGitProviderConfigStore.On("ListConfigsForUrl", pendingBuild.Repository.Url).Return([]*gitprovider.GitProviderConfig{&gitProviderConfig}, nil)
	s.mockGitService.On("CloneRepository", pendingBuild.Repository, &http.BasicAuth{
		Username: gitProviderConfig.Username,
	}).Return(nil)

	mockGitService := git_mocks.NewMockGitService()
	mockGitService.On("CloneRepository", pendingBuild.Repository, &http.BasicAuth{
		Username: gitProviderConfig.Username,
	}).Return(nil)

	runningBuild := *mocks.MockBuild
	runningBuild.State = build.BuildStateRunning
	s.mockBuilder.On("Build", runningBuild).Return("image", "user", nil)

	successBuild := *mocks.MockBuild
	successBuild.State = build.BuildStateSuccess
	successBuild.Image = util.Pointer("image")
	successBuild.User = util.Pointer("user")
	s.mockBuilder.On("Publish", successBuild).Return(nil)

	s.mockBuilder.On("CleanUp").Return(nil)

	mockLogger := logger_mocks.NewMockLogger()
	mockLogger.On("Write", mock.Anything).Return(0, nil)

	s.Runner.RunBuildProcess(build.BuildProcessConfig{
		Builder:      &s.mockBuilder,
		BuildLogger:  mockLogger,
		Build:        mocks.MockBuild,
		WorkspaceDir: "",
		GitService:   mockGitService,
		Wg:           nil,
	})

	mockLogger.AssertExpectations(s.T())
	s.mockBuilder.AssertExpectations(s.T())
	s.mockGitProviderConfigStore.AssertExpectations(s.T())

	s.Require().Equal(mocks.MockBuild.Image, util.Pointer("image"))
	s.Require().Equal(mocks.MockBuild.User, util.Pointer("user"))
	s.Require().Equal(mocks.MockBuild.State, build.BuildStatePublished)
}
