// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider/util"
)

type ProviderRPCClient struct {
	client *rpc.Client
}

func (m *ProviderRPCClient) Initialize(req InitializeProviderRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.Initialize", &req, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) GetInfo() (models.ProviderInfo, error) {
	var resp models.ProviderInfo
	err := m.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	return resp, err
}

func (m *ProviderRPCClient) CheckRequirements() (*[]RequirementStatus, error) {
	var result []RequirementStatus
	err := m.client.Call("Plugin.CheckRequirements", new(interface{}), &result)
	return &result, err
}

func (m *ProviderRPCClient) GetPresetTargetConfigs() (*[]TargetConfig, error) {
	var resp []TargetConfig
	err := m.client.Call("Plugin.GetPresetTargetConfigs", new(interface{}), &resp)
	return &resp, err
}

func (m *ProviderRPCClient) CreateTarget(targetReq *TargetRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.CreateTarget", targetReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) StartTarget(targetReq *TargetRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.StartTarget", targetReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) StopTarget(targetReq *TargetRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.StopTarget", targetReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) DestroyTarget(targetReq *TargetRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.DestroyTarget", targetReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) GetTargetProviderMetadata(targetReq *TargetRequest) (string, error) {
	var resp string
	err := m.client.Call("Plugin.GetTargetProviderMetadata", targetReq, &resp)
	return resp, err
}

func (m *ProviderRPCClient) CreateWorkspace(workspaceReq *WorkspaceRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.CreateWorkspace", workspaceReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) StartWorkspace(workspaceReq *WorkspaceRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.StartWorkspace", workspaceReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) StopWorkspace(workspaceReq *WorkspaceRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.StopWorkspace", workspaceReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) DestroyWorkspace(workspaceReq *WorkspaceRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.DestroyWorkspace", workspaceReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) GetWorkspaceProviderMetadata(workspaceReq *WorkspaceRequest) (string, error) {
	var resp string
	err := m.client.Call("Plugin.GetWorkspaceProviderMetadata", workspaceReq, &resp)
	return resp, err
}
