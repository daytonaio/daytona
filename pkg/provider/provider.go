// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/hashicorp/go-plugin"
)

type Provider interface {
	Initialize(InitializeProviderRequest) (*util.Empty, error)
	GetInfo() (models.ProviderInfo, error)
	CheckRequirements() (*[]RequirementStatus, error)

	GetPresetTargetConfigs() (*[]TargetConfig, error)

	CreateTarget(*TargetRequest) (*util.Empty, error)
	StartTarget(*TargetRequest) (*util.Empty, error)
	StopTarget(*TargetRequest) (*util.Empty, error)
	DestroyTarget(*TargetRequest) (*util.Empty, error)
	GetTargetProviderMetadata(*TargetRequest) (string, error)

	CreateWorkspace(*WorkspaceRequest) (*util.Empty, error)
	StartWorkspace(*WorkspaceRequest) (*util.Empty, error)
	StopWorkspace(*WorkspaceRequest) (*util.Empty, error)
	DestroyWorkspace(*WorkspaceRequest) (*util.Empty, error)
	GetWorkspaceProviderMetadata(*WorkspaceRequest) (string, error)
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
