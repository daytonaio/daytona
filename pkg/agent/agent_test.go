// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_test

import (
	"bytes"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daytonaio/daytona/internal/testing/agent/mocks"
	mock_git "github.com/daytonaio/daytona/internal/testing/git/mocks"
	"github.com/daytonaio/daytona/pkg/agent"
	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
)

var workspace1 = &workspace.Workspace{
	Id:   "123",
	Name: "test",
	Repository: &gitprovider.GitRepository{
		Id:   "123",
		Url:  "https://github.com/daytonaio/daytona",
		Name: "daytona",
	},
	TargetId: "123",
	State: &workspace.WorkspaceState{
		UpdatedAt: "123",
		Uptime:    148,
		GitStatus: gitStatus1,
	},
}

var target1 = &target.Target{
	Id:           "123",
	Name:         "test",
	TargetConfig: "local",
}

var gitStatus1 = &workspace.GitStatus{
	CurrentBranch: "main",
	Files: []*workspace.FileStatus{{
		Name:     "File1",
		Extra:    "",
		Staging:  workspace.Modified,
		Worktree: workspace.Modified,
	}},
}

var mockConfig = &config.Config{
	TargetId:    target1.Id,
	WorkspaceId: workspace1.Id,
	Server: config.DaytonaServerConfig{
		ApiKey: "test-api-key",
	},
	Mode: config.ModeWorkspace,
}

func TestAgent(t *testing.T) {
	buf := bytes.Buffer{}
	log.SetOutput(&buf)

	apiServer := mocks.NewMockRestServer(t)
	defer apiServer.Close()

	mockConfig.Server.ApiUrl = apiServer.URL

	mockGitService := mock_git.NewMockGitService()
	mockGitService.On("RepositoryExists").Return(true, nil)
	mockGitService.On("SetGitConfig", mock.Anything, mock.Anything).Return(nil)
	mockGitService.On("GetGitStatus").Return(gitStatus1, nil)

	mockSshServer := mocks.NewMockSshServer()
	mockTailscaleServer := mocks.NewMockTailscaleServer()

	mockConfig.WorkspaceDir = t.TempDir()

	// Create a new Agent instance
	a := &agent.Agent{
		Config:    mockConfig,
		Git:       mockGitService,
		Ssh:       mockSshServer,
		Tailscale: mockTailscaleServer,
		Workspace: workspace1,
	}

	t.Run("Start agent", func(t *testing.T) {
		err := a.Start()

		require.Nil(t, err)
	})

	t.Cleanup(func() {
		mockGitService.AssertExpectations(t)
		mockSshServer.AssertExpectations(t)
		mockTailscaleServer.AssertExpectations(t)
	})
}

func TestAgentHostMode(t *testing.T) {
	mockGitService := mock_git.NewMockGitService()
	mockSshServer := mocks.NewMockSshServer()
	mockTailscaleServer := mocks.NewMockTailscaleServer()

	mockConfig := *mockConfig
	mockConfig.Mode = config.ModeHost

	// Create a new Agent instance
	a := &agent.Agent{
		Config:    &mockConfig,
		Git:       mockGitService,
		Ssh:       mockSshServer,
		Tailscale: mockTailscaleServer,
	}

	t.Run("Start agent in host mode", func(t *testing.T) {
		mockConfig.Mode = config.ModeHost
		err := a.Start()

		require.Nil(t, err)
	})

	t.Cleanup(func() {
		mockGitService.AssertExpectations(t)
		mockSshServer.AssertExpectations(t)
		mockTailscaleServer.AssertExpectations(t)
	})
}
