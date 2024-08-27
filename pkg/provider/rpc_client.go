// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
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
	var response workspace.WorkspaceInfo
	err := m.client.Call("Plugin.GetWorkspaceInfo", workspaceReq, &response)
	return &response, err
}

func (m *ProviderRPCClient) CreateProject(projectReq *ProjectRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.CreateProject", projectReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) StartProject(projectReq *ProjectRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.StartProject", projectReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) StopProject(projectReq *ProjectRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.StopProject", projectReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) DestroyProject(projectReq *ProjectRequest) (*util.Empty, error) {
	err := m.client.Call("Plugin.DestroyProject", projectReq, new(util.Empty))
	return new(util.Empty), err
}

func (m *ProviderRPCClient) GetProjectInfo(projectReq *ProjectRequest) (*project.ProjectInfo, error) {
	var resp project.ProjectInfo
	err := m.client.Call("Plugin.GetProjectInfo", projectReq, &resp)
	return &resp, err
}
