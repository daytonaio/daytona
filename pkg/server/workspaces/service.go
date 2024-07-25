// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type IWorkspaceService interface {
	CreateWorkspace(ctx context.Context, req dto.CreateWorkspaceRequest) (*workspace.Workspace, error)
	GetWorkspace(ctx context.Context, workspaceId string) (*dto.WorkspaceDTO, error)
	GetWorkspaceLogReader(workspaceId string) (io.Reader, error)
	GetProjectLogReader(workspaceId, projectName string) (io.Reader, error)
	ListWorkspaces(verbose bool) ([]dto.WorkspaceDTO, error)
	RemoveWorkspace(ctx context.Context, workspaceId string) error
	ForceRemoveWorkspace(ctx context.Context, workspaceId string) error
	SetProjectState(workspaceId string, projectName string, state *workspace.ProjectState) (*workspace.Workspace, error)
	StartProject(ctx context.Context, workspaceId string, projectName string) error
	StartWorkspace(ctx context.Context, workspaceId string) error
	StopProject(ctx context.Context, workspaceId string, projectName string) error
	StopWorkspace(ctx context.Context, workspaceId string) error
}

type targetStore interface {
	Find(targetName string) (*provider.ProviderTarget, error)
}

type WorkspaceServiceConfig struct {
	WorkspaceStore           workspace.Store
	TargetStore              targetStore
	ContainerRegistryService containerregistries.IContainerRegistryService
	ServerApiUrl             string
	ServerUrl                string
	Provisioner              provisioner.IProvisioner
	DefaultProjectImage      string
	DefaultProjectUser       string
	ApiKeyService            apikeys.IApiKeyService
	LoggerFactory            logs.LoggerFactory
	GitProviderService       gitproviders.IGitProviderService
	BuilderFactory           build.IBuilderFactory
	TelemetryService         telemetry.TelemetryService
}

func NewWorkspaceService(config WorkspaceServiceConfig) IWorkspaceService {
	return &WorkspaceService{
		workspaceStore:           config.WorkspaceStore,
		targetStore:              config.TargetStore,
		containerRegistryService: config.ContainerRegistryService,
		serverApiUrl:             config.ServerApiUrl,
		serverUrl:                config.ServerUrl,
		defaultProjectImage:      config.DefaultProjectImage,
		defaultProjectUser:       config.DefaultProjectUser,
		provisioner:              config.Provisioner,
		loggerFactory:            config.LoggerFactory,
		apiKeyService:            config.ApiKeyService,
		gitProviderService:       config.GitProviderService,
		builderFactory:           config.BuilderFactory,
		telemetryService:         config.TelemetryService,
	}
}

type WorkspaceService struct {
	workspaceStore           workspace.Store
	targetStore              targetStore
	containerRegistryService containerregistries.IContainerRegistryService
	provisioner              provisioner.IProvisioner
	apiKeyService            apikeys.IApiKeyService
	serverApiUrl             string
	serverUrl                string
	defaultProjectImage      string
	defaultProjectUser       string
	loggerFactory            logs.LoggerFactory
	gitProviderService       gitproviders.IGitProviderService
	builderFactory           build.IBuilderFactory
	telemetryService         telemetry.TelemetryService
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
