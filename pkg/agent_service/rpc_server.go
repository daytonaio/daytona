// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_service

import (
	"github.com/daytonaio/daytona/pkg/types"
)

type AgentServiceRPCServer struct {
	Impl AgentService
}

func (m *AgentServiceRPCServer) Initialize(arg InitializeAgentServiceRequest) error {
	_, err := m.Impl.Initialize(arg)
	return err
}

func (m *AgentServiceRPCServer) GetInfo(arg interface{}, resp *AgentServiceInfo) error {
	info, err := m.Impl.GetInfo()
	if err != nil {
		return err
	}

	*resp = info
	return nil
}

func (m *AgentServiceRPCServer) SetConfig(arg *AgentServiceConfig, resp *types.Empty) error {
	_, err := m.Impl.SetConfig(arg)
	return err
}

func (m *AgentServiceRPCServer) ProjectPreInit(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.ProjectPreInit(arg)
	return err
}

func (m *AgentServiceRPCServer) ProjectPostInit(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.ProjectPostInit(arg)
	return err
}

func (m *AgentServiceRPCServer) ProjectPreStart(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.ProjectPreStart(arg)
	return err
}

func (m *AgentServiceRPCServer) ProjectPostStart(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.ProjectPostStart(arg)
	return err
}

func (m *AgentServiceRPCServer) ProjectPreStop(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.ProjectPreStop(arg)
	return err
}

func (m *AgentServiceRPCServer) GetProjectInfo(arg *types.Project, resp *types.ProjectInfo) error {
	info, err := m.Impl.GetProjectInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}

func (m *AgentServiceRPCServer) LivenessProbe(arg interface{}, resp *types.Empty) error {
	_, err := m.Impl.LivenessProbe()
	return err
}

func (m *AgentServiceRPCServer) LivenessProbeTimeout(arg interface{}, resp *uint32) error {
	timeout := m.Impl.LivenessProbeTimeout()
	*resp = timeout
	return nil
}
