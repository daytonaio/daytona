// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider/util"
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

func (m *ProviderRPCServer) GetTargetConfigManifest(arg interface{}, resp *TargetConfigManifest) error {
	targetConfigManifest, err := m.Impl.GetTargetConfigManifest()
	if err != nil {
		return err
	}

	*resp = *targetConfigManifest
	return nil
}

func (m *ProviderRPCServer) GetPresetTargetConfigs(arg interface{}, resp *[]TargetConfig) error {
	targetConfigs, err := m.Impl.GetPresetTargetConfigs()
	if err != nil {
		return err
	}

	*resp = *targetConfigs
	return nil
}

func (m *ProviderRPCServer) CreateTarget(arg *TargetRequest, resp *util.Empty) error {
	_, err := m.Impl.CreateTarget(arg)
	return err
}

func (m *ProviderRPCServer) StartTarget(arg *TargetRequest, resp *util.Empty) error {
	_, err := m.Impl.StartTarget(arg)
	return err
}

func (m *ProviderRPCServer) StopTarget(arg *TargetRequest, resp *util.Empty) error {
	_, err := m.Impl.StopTarget(arg)
	return err
}

func (m *ProviderRPCServer) DestroyTarget(arg *TargetRequest, resp *util.Empty) error {
	_, err := m.Impl.DestroyTarget(arg)
	return err
}

func (m *ProviderRPCServer) GetTargetInfo(arg *TargetRequest, resp *models.TargetInfo) error {
	info, err := m.Impl.GetTargetInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
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

func (m *ProviderRPCServer) GetWorkspaceInfo(arg *WorkspaceRequest, resp *models.WorkspaceInfo) error {
	info, err := m.Impl.GetWorkspaceInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}
