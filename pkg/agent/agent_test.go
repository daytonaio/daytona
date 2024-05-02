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
	"github.com/daytonaio/daytona/pkg/workspace"
)

var project1 = &workspace.Project{
	Name: "test",
	Repository: &gitprovider.GitRepository{
		Id:   "123",
		Url:  "https://github.com/daytonaio/daytona",
		Name: "daytona",
	},
	WorkspaceId:       "123",
	Target:            "local",
	PostStartCommands: []string{"echo 'test' > test.txt"},
}

var workspace1 = &workspace.Workspace{
	Id:     "123",
	Name:   "test",
	Target: "local",
	Projects: []*workspace.Project{
		project1,
	},
}

var mockConfig = &config.Config{
	WorkspaceId: workspace1.Id,
	ProjectName: project1.Name,
	Server: config.DaytonaServerConfig{
		ApiKey: "test-api-key",
	},
	Mode: config.ModeProject,
}

func TestAgent(t *testing.T) {
	buf := bytes.Buffer{}
	log.SetOutput(&buf)

	apiServer := mocks.NewMockRestServer(t, workspace1)
	defer apiServer.Close()

	mockConfig.Server.ApiUrl = apiServer.URL

	mockGitService := mock_git.NewMockGitService()
	mockGitService.On("RepositoryExists", project1).Return(true, nil)
	mockGitService.On("SetGitConfig", mock.Anything).Return(nil)

	mockSshServer := mocks.NewMockSshServer()
	mockTailscaleServer := mocks.NewMockTailscaleServer()

	mockConfig.ProjectDir = t.TempDir()

	// Create a new Agent instance
	a := &agent.Agent{
		Config:    mockConfig,
		Git:       mockGitService,
		Ssh:       mockSshServer,
		Tailscale: mockTailscaleServer,
	}

	t.Run("Start agent", func(t *testing.T) {
		err := a.Start()

		require.Nil(t, err)

		// Check post start command result
		require.FileExists(t, mockConfig.ProjectDir+"/test.txt")
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
