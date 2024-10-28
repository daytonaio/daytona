// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfig"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type ITargetService interface {
	CreateTarget(ctx context.Context, req dto.CreateTargetDTO) (*target.Target, error)
	GetTarget(ctx context.Context, targetId string, verbose bool) (*dto.TargetDTO, error)
	GetTargetLogReader(targetId string) (io.Reader, error)
	ListTargets(ctx context.Context, verbose bool) ([]dto.TargetDTO, error)
	StartTarget(ctx context.Context, targetId string) error
	StopTarget(ctx context.Context, targetId string) error
	RemoveTarget(ctx context.Context, targetId string) error
	ForceRemoveTarget(ctx context.Context, targetId string) error

	GetWorkspaceLogReader(targetId, workspaceName string) (io.Reader, error)
	SetWorkspaceState(targetId string, workspaceName string, state *workspace.WorkspaceState) (*target.Target, error)
	StartWorkspace(ctx context.Context, targetId string, workspaceName string) error
	StopWorkspace(ctx context.Context, targetId string, workspaceName string) error
}

type targetConfigStore interface {
	Find(filter *provider.TargetConfigFilter) (*provider.TargetConfig, error)
}

type TargetServiceConfig struct {
	TargetStore              target.Store
	TargetConfigStore        targetConfigStore
	ContainerRegistryService containerregistries.IContainerRegistryService
	BuildService             builds.IBuildService
	WorkspaceConfigService   workspaceconfig.IWorkspaceConfigService
	ServerApiUrl             string
	ServerUrl                string
	ServerVersion            string
	Provisioner              provisioner.IProvisioner
	DefaultWorkspaceImage    string
	DefaultWorkspaceUser     string
	BuilderImage             string
	ApiKeyService            apikeys.IApiKeyService
	LoggerFactory            logs.LoggerFactory
	GitProviderService       gitproviders.IGitProviderService
	TelemetryService         telemetry.TelemetryService
}

func NewTargetService(config TargetServiceConfig) ITargetService {
	return &TargetService{
		targetStore:              config.TargetStore,
		targetConfigStore:        config.TargetConfigStore,
		containerRegistryService: config.ContainerRegistryService,
		buildService:             config.BuildService,
		workspaceConfigService:   config.WorkspaceConfigService,
		serverApiUrl:             config.ServerApiUrl,
		serverUrl:                config.ServerUrl,
		serverVersion:            config.ServerVersion,
		defaultWorkspaceImage:    config.DefaultWorkspaceImage,
		defaultWorkspaceUser:     config.DefaultWorkspaceUser,
		provisioner:              config.Provisioner,
		loggerFactory:            config.LoggerFactory,
		apiKeyService:            config.ApiKeyService,
		gitProviderService:       config.GitProviderService,
		telemetryService:         config.TelemetryService,
		builderImage:             config.BuilderImage,
	}
}

type TargetService struct {
	targetStore              target.Store
	targetConfigStore        targetConfigStore
	containerRegistryService containerregistries.IContainerRegistryService
	buildService             builds.IBuildService
	workspaceConfigService   workspaceconfig.IWorkspaceConfigService
	provisioner              provisioner.IProvisioner
	apiKeyService            apikeys.IApiKeyService
	serverApiUrl             string
	serverUrl                string
	serverVersion            string
	defaultWorkspaceImage    string
	defaultWorkspaceUser     string
	builderImage             string
	loggerFactory            logs.LoggerFactory
	gitProviderService       gitproviders.IGitProviderService
	telemetryService         telemetry.TelemetryService
}

func (s *TargetService) SetWorkspaceState(targetId, workspaceName string, state *workspace.WorkspaceState) (*target.Target, error) {
	tg, err := s.targetStore.Find(targetId)
	if err != nil {
		return nil, err
	}

	for _, workspace := range tg.Workspaces {
		if workspace.Name == workspaceName {
			workspace.State = state
			return tg, s.targetStore.Save(tg)
		}
	}

	return nil, errors.New("workspace not found")
}

func (s *TargetService) GetTargetLogReader(targetId string) (io.Reader, error) {
	return s.loggerFactory.CreateTargetLogReader(targetId)
}

func (s *TargetService) GetWorkspaceLogReader(targetId, workspaceName string) (io.Reader, error) {
	return s.loggerFactory.CreateWorkspaceLogReader(targetId, workspaceName)
}
