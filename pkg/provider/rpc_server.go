// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/project"
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

func (m *ProviderRPCServer) CheckRequirements(arg interface{}, resp *[]RequirementStatus) error {
	result, err := m.Impl.CheckRequirements()
	if err != nil {
		return err
	}
	*resp = *result
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

func (m *ProviderRPCServer) GetTargetInfo(arg *TargetRequest, resp *target.TargetInfo) error {
	info, err := m.Impl.GetTargetInfo(arg)
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
