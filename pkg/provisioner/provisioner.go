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

type IProvisioner interface {
	CreateTarget(target *target.Target) error
	StartTarget(target *target.Target) error
	StopTarget(target *target.Target) error
	GetTargetInfo(ctx context.Context, target *target.Target) (*target.TargetInfo, error)
	DestroyTarget(target *target.Target) error

	CreateWorkspace(workspace *workspace.Workspace, target *target.Target, cr *containerregistry.ContainerRegistry, gc *gitprovider.GitProviderConfig) error
	DestroyWorkspace(workspace *workspace.Workspace, target *target.Target) error
	StartWorkspace(workspace *workspace.Workspace, target *target.Target) error
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
