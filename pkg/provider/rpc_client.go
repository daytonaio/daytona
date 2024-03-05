// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/types"
)

type ProviderRPCClient struct {
	client *rpc.Client
}

func (m *ProviderRPCClient) Initialize(req InitializeProviderRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.Initialize", &req, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) GetInfo() (ProviderInfo, error) {
	var resp ProviderInfo
	err := m.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	return resp, err
}

func (m *ProviderRPCClient) GetTargetManifest() (*ProviderTargetManifest, error) {
	var resp ProviderTargetManifest
	err := m.client.Call("Plugin.GetTargetManifest", new(interface{}), &resp)

	return &resp, err
}

func (m *ProviderRPCClient) GetDefaultTargets() (*[]ProviderTarget, error) {
	var resp []ProviderTarget
	err := m.client.Call("Plugin.GetDefaultTargets", new(interface{}), &resp)
	return &resp, err
}

func (m *ProviderRPCClient) CreateWorkspace(workspaceReq *WorkspaceRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.CreateWorkspace", workspaceReq, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) StartWorkspace(workspaceReq *WorkspaceRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.StartWorkspace", workspaceReq, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) StopWorkspace(workspaceReq *WorkspaceRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.StopWorkspace", workspaceReq, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) DestroyWorkspace(workspaceReq *WorkspaceRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.DestroyWorkspace", workspaceReq, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) GetWorkspaceInfo(workspace *WorkspaceRequest) (*types.WorkspaceInfo, error) {
	var response types.WorkspaceInfo
	err := m.client.Call("Plugin.GetWorkspaceInfo", workspace, &response)
	return &response, err
}

func (m *ProviderRPCClient) CreateProject(projectReq *ProjectRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.CreateProject", projectReq, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) StartProject(projectReq *ProjectRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.StartProject", projectReq, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) StopProject(projectReq *ProjectRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.StopProject", projectReq, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) DestroyProject(projectReq *ProjectRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.DestroyProject", projectReq, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) GetProjectInfo(projectReq *ProjectRequest) (*types.ProjectInfo, error) {
	var resp types.ProjectInfo
	err := m.client.Call("Plugin.GetProjectInfo", projectReq, &resp)
	return &resp, err
}
