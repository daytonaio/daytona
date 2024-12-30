// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
)

type ProjectParams struct {
	Project                       *project.Project
	Target                        *provider.ProviderTarget
	ContainerRegistry             *containerregistry.ContainerRegistry
	GitProviderConfig             *gitprovider.GitProviderConfig
	BuilderImage                  string
	BuilderImageContainerRegistry *containerregistry.ContainerRegistry
}

type IProvisioner interface {
	CreateProject(params ProjectParams) error
	CreateWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error
	DestroyProject(project *project.Project, target *provider.ProviderTarget) error
	DestroyWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error
	GetWorkspaceInfo(ctx context.Context, workspace *workspace.Workspace, target *provider.ProviderTarget) (*workspace.WorkspaceInfo, error)
	StartProject(params ProjectParams) error
	StartWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error
	StopProject(project *project.Project, target *provider.ProviderTarget) error
	StopWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error
}

type ProvisionerConfig struct {
	ProviderManager manager.IProviderManager
}

func NewProvisioner(config ProvisionerConfig) IProvisioner {
	return &Provisioner{
		providerManager: config.ProviderManager,
	}
}

type Provisioner struct {
	providerManager manager.IProviderManager
}
