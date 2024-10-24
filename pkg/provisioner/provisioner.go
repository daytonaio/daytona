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

type IProvisioner interface {
	CreateProject(project *project.Project, targetConfig *provider.TargetConfig, cr *containerregistry.ContainerRegistry, gc *gitprovider.GitProviderConfig) error
	CreateWorkspace(workspace *workspace.Workspace, targetConfig *provider.TargetConfig) error
	DestroyProject(project *project.Project, targetConfig *provider.TargetConfig) error
	DestroyWorkspace(workspace *workspace.Workspace, targetConfig *provider.TargetConfig) error
	GetWorkspaceInfo(ctx context.Context, workspace *workspace.Workspace, targetConfig *provider.TargetConfig) (*workspace.WorkspaceInfo, error)
	StartProject(project *project.Project, targetConfig *provider.TargetConfig) error
	StartWorkspace(workspace *workspace.Workspace, targetConfig *provider.TargetConfig) error
	StopProject(project *project.Project, targetConfig *provider.TargetConfig) error
	StopWorkspace(workspace *workspace.Workspace, targetConfig *provider.TargetConfig) error
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
