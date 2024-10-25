// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/project"
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
