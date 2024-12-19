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

func (m *ProviderRPCServer) GetInfo(arg interface{}, resp *models.ProviderInfo) error {
	info, err := m.Impl.GetInfo()
	if err != nil {
		return err
	}

	*resp = info
	return nil
}

func (m *ProviderRPCServer) CheckRequirements(arg interface{}, resp *[]RequirementStatus) error {
	result, err := m.Impl.CheckRequirements()
	if err != nil {
		return err
	}
	*resp = *result
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

func (m *ProviderRPCServer) GetTargetProviderMetadata(arg *TargetRequest, resp *string) error {
	metadata, err := m.Impl.GetTargetProviderMetadata(arg)
	if err != nil {
		return err
	}

	*resp = metadata
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

func (m *ProviderRPCServer) GetWorkspaceProviderMetadata(arg *WorkspaceRequest, resp *string) error {
	metadata, err := m.Impl.GetWorkspaceProviderMetadata(arg)
	if err != nil {
		return err
	}

	*resp = metadata
	return nil
}
