// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider/manager"
)

type WorkspaceParams struct {
	Workspace           *models.Workspace
	Target              *models.Target
	ContainerRegistries common.ContainerRegistries
	GitProviderConfig   *models.GitProviderConfig
	BuilderImage        string
}

type IProvisioner interface {
	CreateTarget(target *models.Target) error
	StartTarget(target *models.Target) error
	StopTarget(target *models.Target) error
	GetTargetInfo(ctx context.Context, target *models.Target) (*models.TargetInfo, error)
	DestroyTarget(target *models.Target) error

	CreateWorkspace(params WorkspaceParams) error
	DestroyWorkspace(workspace *models.Workspace) error
	StartWorkspace(params WorkspaceParams) error
	StopWorkspace(workspace *models.Workspace) error
	GetWorkspaceInfo(ctx context.Context, workspace *models.Workspace) (*models.WorkspaceInfo, error)
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
