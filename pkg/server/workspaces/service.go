// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type targetStore interface {
	Find(targetName string) (*provider.ProviderTarget, error)
}

type WorkspaceServiceConfig struct {
	WorkspaceStore        workspace.Store
	TargetStore           targetStore
	ServerApiUrl          string
	ServerUrl             string
	Provisioner           provisioner.Provisioner
	ApiKeyService         apikeys.ApiKeyService
	NewWorkspaceLogger    func(workspaceId string) logger.Logger
	NewProjectLogger      func(workspaceId, projectName string) logger.Logger
	NewWorkspaceLogReader func(workspaceId string) (io.Reader, error)
}

func NewWorkspaceService(config WorkspaceServiceConfig) *WorkspaceService {
	return &WorkspaceService{
		workspaceStore:        config.WorkspaceStore,
		targetStore:           config.TargetStore,
		serverApiUrl:          config.ServerApiUrl,
		serverUrl:             config.ServerUrl,
		provisioner:           config.Provisioner,
		newWorkspaceLogger:    config.NewWorkspaceLogger,
		newProjectLogger:      config.NewProjectLogger,
		apiKeyService:         config.ApiKeyService,
		NewWorkspaceLogReader: config.NewWorkspaceLogReader,
	}
}

type WorkspaceService struct {
	workspaceStore        workspace.Store
	targetStore           targetStore
	provisioner           provisioner.Provisioner
	apiKeyService         apikeys.ApiKeyService
	serverApiUrl          string
	serverUrl             string
	newWorkspaceLogger    func(workspaceId string) logger.Logger
	newProjectLogger      func(workspaceId, projectName string) logger.Logger
	NewWorkspaceLogReader func(workspaceId string) (io.Reader, error)
}

func (s *WorkspaceService) SetProjectState(workspaceId, projectName string, state *workspace.ProjectState) (*workspace.Workspace, error) {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return nil, err
	}

	for _, project := range ws.Projects {
		if project.Name == projectName {
			project.State = state
			return ws, s.workspaceStore.Save(ws)
		}
	}

	return nil, errors.New("project not found")
}
