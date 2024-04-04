// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/types"
)

type IProvisioner interface {
	CreateWorkspace(workspace *types.Workspace, target *provider.ProviderTarget) error
	CreateProject(project *types.Project, target *provider.ProviderTarget) error
	DestroyWorkspace(workspace *types.Workspace, target *provider.ProviderTarget) error
	DestroyProject(project *types.Project, target *provider.ProviderTarget) error
	StartWorkspace(workspace *types.Workspace, target *provider.ProviderTarget) error
	StopWorkspace(workspace *types.Workspace, target *provider.ProviderTarget) error
	GetWorkspaceInfo(workspace *types.Workspace, target *provider.ProviderTarget) (*types.WorkspaceInfo, error)
}
