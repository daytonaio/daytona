// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/types"
)

type ProviderRPCServer struct {
	Impl Provider
}

func (m *ProviderRPCServer) Initialize(arg InitializeProviderRequest, resp *types.Empty) error {
	_, err := m.Impl.Initialize(arg)
	return err
}

func (m *ProviderRPCServer) GetInfo(arg interface{}, resp *ProviderInfo) error {
	info, err := m.Impl.GetInfo()
	if err != nil {
		return err
	}

	*resp = info
	return nil
}

func (m *ProviderRPCServer) Configure(arg interface{}, configResponse *interface{}) error {
	config, err := m.Impl.Configure()
	if err != nil {
		return err
	}

	*configResponse = config
	return nil
}

func (m *ProviderRPCServer) CreateWorkspace(arg *types.Workspace, resp *types.Empty) error {
	_, err := m.Impl.CreateWorkspace(arg)
	return err
}

func (m *ProviderRPCServer) StartWorkspace(arg *types.Workspace, resp *types.Empty) error {
	_, err := m.Impl.StartWorkspace(arg)
	return err
}

func (m *ProviderRPCServer) StopWorkspace(arg *types.Workspace, resp *types.Empty) error {
	_, err := m.Impl.StopWorkspace(arg)
	return err
}

func (m *ProviderRPCServer) DestroyWorkspace(arg *types.Workspace, resp *types.Empty) error {
	_, err := m.Impl.DestroyWorkspace(arg)
	return err
}

func (m *ProviderRPCServer) GetWorkspaceInfo(arg *types.Workspace, resp *types.WorkspaceInfo) error {
	info, err := m.Impl.GetWorkspaceInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}

func (m *ProviderRPCServer) CreateProject(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.CreateProject(arg)
	return err
}

func (m *ProviderRPCServer) StartProject(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.StartProject(arg)
	return err
}

func (m *ProviderRPCServer) StopProject(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.StopProject(arg)
	return err
}

func (m *ProviderRPCServer) DestroyProject(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.DestroyProject(arg)
	return err
}

func (m *ProviderRPCServer) GetProjectInfo(arg *types.Project, resp *types.ProjectInfo) error {
	info, err := m.Impl.GetProjectInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}
