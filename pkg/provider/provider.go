// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/project"
	"github.com/hashicorp/go-plugin"
)

type Provider interface {
	Initialize(InitializeProviderRequest) (*util.Empty, error)
	GetInfo() (ProviderInfo, error)
	CheckRequirements() (*[]RequirementStatus, error)

	GetTargetConfigManifest() (*TargetConfigManifest, error)
	GetPresetTargetConfigs() (*[]TargetConfig, error)

	CreateTarget(*TargetRequest) (*util.Empty, error)
	StartTarget(*TargetRequest) (*util.Empty, error)
	StopTarget(*TargetRequest) (*util.Empty, error)
	DestroyTarget(*TargetRequest) (*util.Empty, error)
	GetTargetInfo(*TargetRequest) (*target.TargetInfo, error)

	CreateProject(*ProjectRequest) (*util.Empty, error)
	StartProject(*ProjectRequest) (*util.Empty, error)
	StopProject(*ProjectRequest) (*util.Empty, error)
	DestroyProject(*ProjectRequest) (*util.Empty, error)
	GetProjectInfo(*ProjectRequest) (*project.ProjectInfo, error)
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
