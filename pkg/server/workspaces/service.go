// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type IWorkspaceService interface {
	CreateWorkspace(id string, name string, targetName string, repositories []gitprovider.GitRepository) (*workspace.Workspace, error)
	GetWorkspace(workspaceId string) (*dto.WorkspaceDTO, error)
	GetWorkspaceLogReader(workspaceId string) (io.Reader, error)
	ListWorkspaces(verbose bool) ([]dto.WorkspaceDTO, error)
	RemoveWorkspace(workspaceId string) error
	SetProjectState(workspaceId string, projectName string, state *workspace.ProjectState) (*workspace.Workspace, error)
	StartProject(workspaceId string, projectName string) error
	StartWorkspace(workspaceId string) error
	StopProject(workspaceId string, projectName string) error
	StopWorkspace(workspaceId string) error
}

type targetStore interface {
	Find(targetName string) (*provider.ProviderTarget, error)
}

type WorkspaceServiceConfig struct {
	WorkspaceStore        workspace.Store
	TargetStore           targetStore
	ServerApiUrl          string
	ServerUrl             string
	Provisioner           provisioner.IProvisioner
	ApiKeyService         apikeys.IApiKeyService
	NewWorkspaceLogger    func(workspaceId string) logger.Logger
	NewProjectLogger      func(workspaceId, projectName string) logger.Logger
	NewWorkspaceLogReader func(workspaceId string) (io.Reader, error)
}

func NewWorkspaceService(config WorkspaceServiceConfig) IWorkspaceService {
	return &WorkspaceService{
		workspaceStore:        config.WorkspaceStore,
		targetStore:           config.TargetStore,
		serverApiUrl:          config.ServerApiUrl,
		serverUrl:             config.ServerUrl,
		provisioner:           config.Provisioner,
		newWorkspaceLogger:    config.NewWorkspaceLogger,
		newProjectLogger:      config.NewProjectLogger,
		apiKeyService:         config.ApiKeyService,
		newWorkspaceLogReader: config.NewWorkspaceLogReader,
	}
}

type WorkspaceService struct {
	workspaceStore        workspace.Store
	targetStore           targetStore
	provisioner           provisioner.IProvisioner
	apiKeyService         apikeys.IApiKeyService
	serverApiUrl          string
	serverUrl             string
	newWorkspaceLogger    func(workspaceId string) logger.Logger
	newProjectLogger      func(workspaceId, projectName string) logger.Logger
	newWorkspaceLogReader func(workspaceId string) (io.Reader, error)
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

func (s *WorkspaceService) GetWorkspaceLogReader(workspaceId string) (io.Reader, error) {
	return s.newWorkspaceLogReader(workspaceId)
}
