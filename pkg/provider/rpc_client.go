// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type ProviderRPCClient struct {
	client *rpc.Client
}

func (m *ProviderRPCClient) Initialize(req InitializeProviderRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.Initialize", &req, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) GetInfo() (ProviderInfo, error) {
	var resp ProviderInfo
	err := m.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	return resp, err
}

func (m *ProviderRPCClient) CheckRequirements() (*[]RequirementStatus, error) {
	var result []RequirementStatus
	err := m.client.Call("Plugin.CheckRequirements", new(interface{}), &result)
	return &result, err
}

func (m *ProviderRPCClient) GetTargetConfigManifest() (*TargetConfigManifest, error) {
	var resp TargetConfigManifest
	err := m.client.Call("Plugin.GetTargetConfigManifest", new(interface{}), &resp)

	return &resp, err
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

func (m *ProviderRPCClient) GetTargetInfo(targetReq *TargetRequest) (*target.TargetInfo, error) {
	var response target.TargetInfo
	err := m.client.Call("Plugin.GetTargetInfo", targetReq, &response)
	return &response, err
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

func (m *ProviderRPCClient) GetWorkspaceInfo(workspaceReq *WorkspaceRequest) (*workspace.WorkspaceInfo, error) {
	var resp workspace.WorkspaceInfo
	err := m.client.Call("Plugin.GetWorkspaceInfo", workspaceReq, &resp)
	return &resp, err
}
