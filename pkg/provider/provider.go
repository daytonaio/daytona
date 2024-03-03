// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/types"
	"github.com/hashicorp/go-plugin"
)

type Provider interface {
	Initialize(InitializeProviderRequest) (*types.Empty, error)
	GetInfo() (ProviderInfo, error)

	GetTargetManifest() (*ProviderTargetManifest, error)
	SetTarget(ProviderTarget) (*types.Empty, error)
	RemoveTarget(string) (*types.Empty, error)
	GetTargets() (*[]ProviderTarget, error)

	CreateWorkspace(workspace *types.Workspace) (*types.Empty, error)
	StartWorkspace(workspace *types.Workspace) (*types.Empty, error)
	StopWorkspace(workspace *types.Workspace) (*types.Empty, error)
	DestroyWorkspace(workspace *types.Workspace) (*types.Empty, error)
	GetWorkspaceInfo(workspace *types.Workspace) (*types.WorkspaceInfo, error)

	CreateProject(project *types.Project) (*types.Empty, error)
	StartProject(project *types.Project) (*types.Empty, error)
	StopProject(project *types.Project) (*types.Empty, error)
	DestroyProject(project *types.Project) (*types.Empty, error)
	GetProjectInfo(project *types.Project) (*types.ProjectInfo, error)
}

type ProviderPlugin struct {
	Impl Provider
}

func (p *ProviderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ProviderRPCServer{Impl: p.Impl}, nil
}

func (p *ProviderPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ProviderRPCClient{client: c}, nil
}
