package agent_service

import (
	"net/rpc"

	"github.com/daytonaio/daytona/common/types"
)

type AgentServiceRPCClient struct{ client *rpc.Client }

func (m *AgentServiceRPCClient) Initialize(req InitializeAgentServiceRequest) error {
	return m.client.Call("Plugin.Initialize", req, nil)
}

func (m *AgentServiceRPCClient) GetInfo() (AgentServiceInfo, error) {
	var resp AgentServiceInfo
	err := m.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	return resp, err
}

func (m *AgentServiceRPCClient) SetConfig(config *AgentServiceConfig) error {
	return m.client.Call("Plugin.SetConfig", config, nil)
}

func (m *AgentServiceRPCClient) ProjectPreInit(project *types.Project) error {
	return m.client.Call("Plugin.ProjectPreInit", project, nil)
}

func (m *AgentServiceRPCClient) ProjectPostInit(project *types.Project) error {
	return m.client.Call("Plugin.ProjectPostInit", project, nil)
}

func (m *AgentServiceRPCClient) ProjectPreStart(project *types.Project) error {
	return m.client.Call("Plugin.ProjectPreStart", project, nil)
}

func (m *AgentServiceRPCClient) ProjectPostStart(project *types.Project) error {
	return m.client.Call("Plugin.ProjectPostStart", project, nil)
}

func (m *AgentServiceRPCClient) ProjectPreStop(project *types.Project) error {
	return m.client.Call("Plugin.ProjectPreStop", project, nil)
}

func (m *AgentServiceRPCClient) GetProjectInfo(project *types.Project) (*types.ProjectInfo, error) {
	var resp types.ProjectInfo
	err := m.client.Call("Plugin.GetProjectInfo", project, &resp)
	return &resp, err
}

func (m *AgentServiceRPCClient) LivenessProbe() error {
	return m.client.Call("Plugin.LivenessProbe", new(interface{}), nil)
}

func (m *AgentServiceRPCClient) LivenessProbeTimeout() uint32 {
	var resp uint32
	m.client.Call("Plugin.LivenessProbeTimeout", new(interface{}), &resp)
	return resp
}
