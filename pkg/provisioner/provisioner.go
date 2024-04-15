// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type IProvisioner interface {
	CreateProject(project *workspace.Project, target *provider.ProviderTarget) error
	CreateWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error
	DestroyProject(project *workspace.Project, target *provider.ProviderTarget) error
	DestroyWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error
	GetWorkspaceInfo(workspace *workspace.Workspace, target *provider.ProviderTarget) (*workspace.WorkspaceInfo, error)
	StartProject(project *workspace.Project, target *provider.ProviderTarget) error
	StartWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error
	StopProject(project *workspace.Project, target *provider.ProviderTarget) error
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
