package agent_service

import (
	"github.com/daytonaio/daytona/common/types"
)

type AgentServiceRPCServer struct {
	Impl AgentService
}

func (m *AgentServiceRPCServer) Initialize(arg InitializeAgentServiceRequest) error {
	return m.Impl.Initialize(arg)
}

func (m *AgentServiceRPCServer) GetInfo(arg interface{}, resp *AgentServiceInfo) error {
	info, err := m.Impl.GetInfo()
	if err != nil {
		return err
	}

	*resp = info
	return nil
}

func (m *AgentServiceRPCServer) SetConfig(arg *AgentServiceConfig, resp interface{}) error {
	return m.Impl.SetConfig(arg)
}

func (m *AgentServiceRPCServer) ProjectPreInit(arg *types.Project, resp interface{}) error {
	return m.Impl.ProjectPreInit(arg)
}

func (m *AgentServiceRPCServer) ProjectPostInit(arg *types.Project, resp interface{}) error {
	return m.Impl.ProjectPostInit(arg)
}

func (m *AgentServiceRPCServer) ProjectPreStart(arg *types.Project, resp interface{}) error {
	return m.Impl.ProjectPreStart(arg)
}

func (m *AgentServiceRPCServer) ProjectPostStart(arg *types.Project, resp interface{}) error {
	return m.Impl.ProjectPostStart(arg)
}

func (m *AgentServiceRPCServer) ProjectPreStop(arg *types.Project, resp interface{}) error {
	return m.Impl.ProjectPreStop(arg)
}

func (m *AgentServiceRPCServer) GetProjectInfo(arg *types.Project, resp *types.ProjectInfo) error {
	info, err := m.Impl.GetProjectInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}

func (m *AgentServiceRPCServer) LivenessProbe(arg interface{}, resp interface{}) error {
	return m.Impl.LivenessProbe()
}

func (m *AgentServiceRPCServer) LivenessProbeTimeout(arg interface{}, resp *uint32) error {
	timeout := m.Impl.LivenessProbeTimeout()
	*resp = timeout
	return nil
}
