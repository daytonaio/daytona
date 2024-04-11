// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_test

import (
	"bytes"
	"net/http"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/daytonaio/daytona/internal/testing/agent/mocks"
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
	WorkspaceId: "123",
	Target:      "local",
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
		Url:    "http://localhost:3000",
		ApiKey: "test-api-key",
	},
}

func TestAgent(t *testing.T) {
	buf := bytes.Buffer{}
	log.SetOutput(&buf)

	apiServer := mocks.NewMockRestServer(t, workspace1)
	defer apiServer.Close()
	go func() {
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	mockGitService := mocks.NewMockGitService(true, project1)
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
	})

	t.Cleanup(func() {
		mockGitService.AssertExpectations(t)
		mockSshServer.AssertExpectations(t)
		mockTailscaleServer.AssertExpectations(t)
	})
}
