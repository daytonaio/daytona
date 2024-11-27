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
	"github.com/daytonaio/daytona/pkg/models"
)

var workspace1 = &models.Workspace{
	Id:   "123",
	Name: "test",
	Repository: &gitprovider.GitRepository{
		Id:   "123",
		Url:  "https://github.com/daytonaio/daytona",
		Name: "daytona",
	},
	TargetId: "123",
	Metadata: &models.WorkspaceMetadata{
		Uptime:    148,
		GitStatus: gitStatus1,
	},
}

var target1 = &models.Target{
	Id:   "123",
	Name: "test",
	ProviderInfo: models.ProviderInfo{
		Name:    "test-provider",
		Version: "test",
	},
	Options: "test-options",
}

var gitStatus1 = &models.GitStatus{
	CurrentBranch: "main",
	Files: []*models.FileStatus{{
		Name:     "File1",
		Extra:    "",
		Staging:  models.Modified,
		Worktree: models.Modified,
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

func TestAgentTargetMode(t *testing.T) {
	mockGitService := mock_git.NewMockGitService()
	mockSshServer := mocks.NewMockSshServer()
	mockTailscaleServer := mocks.NewMockTailscaleServer()

	mockConfig := *mockConfig
	mockConfig.Mode = config.ModeTarget

	// Create a new Agent instance
	a := &agent.Agent{
		Config:    &mockConfig,
		Git:       mockGitService,
		Ssh:       mockSshServer,
		Tailscale: mockTailscaleServer,
	}

	t.Run("Start agent in target mode", func(t *testing.T) {
		mockConfig.Mode = config.ModeTarget
		err := a.Start()

		require.Nil(t, err)
	})

	t.Cleanup(func() {
		mockGitService.AssertExpectations(t)
		mockSshServer.AssertExpectations(t)
		mockTailscaleServer.AssertExpectations(t)
	})
}
