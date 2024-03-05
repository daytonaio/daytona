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
	GetDefaultTargets() (*[]ProviderTarget, error)

	CreateWorkspace(*WorkspaceRequest) (*types.Empty, error)
	StartWorkspace(*WorkspaceRequest) (*types.Empty, error)
	StopWorkspace(*WorkspaceRequest) (*types.Empty, error)
	DestroyWorkspace(*WorkspaceRequest) (*types.Empty, error)
	GetWorkspaceInfo(*WorkspaceRequest) (*types.WorkspaceInfo, error)

	CreateProject(*ProjectRequest) (*types.Empty, error)
	StartProject(*ProjectRequest) (*types.Empty, error)
	StopProject(*ProjectRequest) (*types.Empty, error)
	DestroyProject(*ProjectRequest) (*types.Empty, error)
	GetProjectInfo(*ProjectRequest) (*types.ProjectInfo, error)
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
