// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/builder"
	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type IWorkspaceService interface {
	CreateWorkspace(req dto.CreateWorkspaceRequest) (*workspace.Workspace, error)
	GetWorkspace(workspaceId string) (*dto.WorkspaceDTO, error)
	GetWorkspaceLogReader(workspaceId string) (io.Reader, error)
	GetProjectLogReader(workspaceId, projectName string) (io.Reader, error)
	ListWorkspaces(verbose bool) ([]dto.WorkspaceDTO, error)
	RemoveWorkspace(workspaceId string) error
	ForceRemoveWorkspace(workspaceId string) error
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
	WorkspaceStore                  workspace.Store
	TargetStore                     targetStore
	ContainerRegistryService        containerregistries.IContainerRegistryService
	ServerApiUrl                    string
	ServerUrl                       string
	Provisioner                     provisioner.IProvisioner
	DefaultProjectImage             string
	DefaultProjectUser              string
	DefaultProjectPostStartCommands []string
	ApiKeyService                   apikeys.IApiKeyService
	LoggerFactory                   logger.LoggerFactory
	GitProviderService              gitproviders.IGitProviderService
	BuilderFactory                  builder.IBuilderFactory
}

func NewWorkspaceService(config WorkspaceServiceConfig) IWorkspaceService {
	return &WorkspaceService{
		workspaceStore:                  config.WorkspaceStore,
		targetStore:                     config.TargetStore,
		containerRegistryService:        config.ContainerRegistryService,
		serverApiUrl:                    config.ServerApiUrl,
		serverUrl:                       config.ServerUrl,
		defaultProjectImage:             config.DefaultProjectImage,
		defaultProjectUser:              config.DefaultProjectUser,
		defaultProjectPostStartCommands: config.DefaultProjectPostStartCommands,
		provisioner:                     config.Provisioner,
		loggerFactory:                   config.LoggerFactory,
		apiKeyService:                   config.ApiKeyService,
		gitProviderService:              config.GitProviderService,
		builderFactory:                  config.BuilderFactory,
	}
}

type WorkspaceService struct {
	workspaceStore                  workspace.Store
	targetStore                     targetStore
	containerRegistryService        containerregistries.IContainerRegistryService
	provisioner                     provisioner.IProvisioner
	apiKeyService                   apikeys.IApiKeyService
	serverApiUrl                    string
	serverUrl                       string
	defaultProjectImage             string
	defaultProjectUser              string
	defaultProjectPostStartCommands []string
	loggerFactory                   logger.LoggerFactory
	gitProviderService              gitproviders.IGitProviderService
	builderFactory                  builder.IBuilderFactory
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
	return s.loggerFactory.CreateWorkspaceLogReader(workspaceId)
}

func (s *WorkspaceService) GetProjectLogReader(workspaceId, projectName string) (io.Reader, error) {
	return s.loggerFactory.CreateProjectLogReader(workspaceId, projectName)
}
