// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type WorkspaceParams struct {
	Workspace                     *workspace.Workspace
	Target                        *target.Target
	ContainerRegistry             *containerregistry.ContainerRegistry
	GitProviderConfig             *gitprovider.GitProviderConfig
	BuilderImage                  string
	BuilderImageContainerRegistry *containerregistry.ContainerRegistry
}

type IProvisioner interface {
	CreateTarget(target *target.Target) error
	StartTarget(target *target.Target) error
	StopTarget(target *target.Target) error
	GetTargetInfo(ctx context.Context, target *target.Target) (*target.TargetInfo, error)
	DestroyTarget(target *target.Target) error

	CreateWorkspace(params WorkspaceParams) error
	DestroyWorkspace(workspace *workspace.Workspace, target *target.Target) error
	StartWorkspace(params WorkspaceParams) error
	StopWorkspace(workspace *workspace.Workspace, target *target.Target) error
	GetWorkspaceInfo(ctx context.Context, workspace *workspace.Workspace, target *target.Target) (*workspace.WorkspaceInfo, error)
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
