// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider/manager"
)

type IProvisioner interface {
	CreateTarget(target *models.Target) error
	StartTarget(target *models.Target) error
	StopTarget(target *models.Target) error
	GetTargetInfo(ctx context.Context, target *models.Target) (*models.TargetInfo, error)
	DestroyTarget(target *models.Target) error

	CreateWorkspace(workspace *models.Workspace, cr *models.ContainerRegistry, gc *models.GitProviderConfig) error
	DestroyWorkspace(workspace *models.Workspace) error
	StartWorkspace(workspace *models.Workspace) error
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
