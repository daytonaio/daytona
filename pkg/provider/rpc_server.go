// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
)

type ProviderRPCServer struct {
	Impl Provider
}

func (m *ProviderRPCServer) Initialize(arg InitializeProviderRequest, resp *util.Empty) error {
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

func (m *ProviderRPCServer) GetTargetManifest(arg interface{}, resp *ProviderTargetManifest) error {
	targetManifest, err := m.Impl.GetTargetManifest()
	if err != nil {
		return err
	}

	*resp = *targetManifest
	return nil
}

func (m *ProviderRPCServer) GetDefaultTargets(arg interface{}, resp *[]ProviderTarget) error {
	targets, err := m.Impl.GetDefaultTargets()
	if err != nil {
		return err
	}

	*resp = *targets
	return nil
}

func (m *ProviderRPCServer) CreateWorkspace(arg *WorkspaceRequest, resp *util.Empty) error {
	_, err := m.Impl.CreateWorkspace(arg)
	return err
}

func (m *ProviderRPCServer) StartWorkspace(arg *WorkspaceRequest, resp *util.Empty) error {
	_, err := m.Impl.StartWorkspace(arg)
	return err
}

func (m *ProviderRPCServer) StopWorkspace(arg *WorkspaceRequest, resp *util.Empty) error {
	_, err := m.Impl.StopWorkspace(arg)
	return err
}

func (m *ProviderRPCServer) DestroyWorkspace(arg *WorkspaceRequest, resp *util.Empty) error {
	_, err := m.Impl.DestroyWorkspace(arg)
	return err
}

func (m *ProviderRPCServer) GetWorkspaceInfo(arg *WorkspaceRequest, resp *workspace.WorkspaceInfo) error {
	info, err := m.Impl.GetWorkspaceInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}

func (m *ProviderRPCServer) CreateProject(arg *ProjectRequest, resp *util.Empty) error {
	_, err := m.Impl.CreateProject(arg)
	return err
}

func (m *ProviderRPCServer) StartProject(arg *ProjectRequest, resp *util.Empty) error {
	_, err := m.Impl.StartProject(arg)
	return err
}

func (m *ProviderRPCServer) StopProject(arg *ProjectRequest, resp *util.Empty) error {
	_, err := m.Impl.StopProject(arg)
	return err
}

func (m *ProviderRPCServer) DestroyProject(arg *ProjectRequest, resp *util.Empty) error {
	_, err := m.Impl.DestroyProject(arg)
	return err
}

func (m *ProviderRPCServer) GetProjectInfo(arg *ProjectRequest, resp *project.ProjectInfo) error {
	info, err := m.Impl.GetProjectInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}
