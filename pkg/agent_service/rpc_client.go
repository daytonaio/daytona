package agent_service

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/types"
)

type AgentServiceRPCClient struct{ client *rpc.Client }

func (m *AgentServiceRPCClient) Initialize(req InitializeAgentServiceRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.Initialize", req, nil)
	return new(types.Empty), err
}

func (m *AgentServiceRPCClient) GetInfo() (AgentServiceInfo, error) {
	var resp AgentServiceInfo
	err := m.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	return resp, err
}

func (m *AgentServiceRPCClient) SetConfig(config *AgentServiceConfig) (*types.Empty, error) {
	err := m.client.Call("Plugin.SetConfig", config, nil)
	return new(types.Empty), err
}

func (m *AgentServiceRPCClient) ProjectPreInit(project *types.Project) (*types.Empty, error) {
	err := m.client.Call("Plugin.ProjectPreInit", project, nil)
	return new(types.Empty), err
}

func (m *AgentServiceRPCClient) ProjectPostInit(project *types.Project) (*types.Empty, error) {
	err := m.client.Call("Plugin.ProjectPostInit", project, nil)
	return new(types.Empty), err
}

func (m *AgentServiceRPCClient) ProjectPreStart(project *types.Project) (*types.Empty, error) {
	err := m.client.Call("Plugin.ProjectPreStart", project, nil)
	return new(types.Empty), err
}

func (m *AgentServiceRPCClient) ProjectPostStart(project *types.Project) (*types.Empty, error) {
	err := m.client.Call("Plugin.ProjectPostStart", project, nil)
	return new(types.Empty), err
}

func (m *AgentServiceRPCClient) ProjectPreStop(project *types.Project) (*types.Empty, error) {
	err := m.client.Call("Plugin.ProjectPreStop", project, nil)
	return new(types.Empty), err
}

func (m *AgentServiceRPCClient) GetProjectInfo(project *types.Project) (*types.ProjectInfo, error) {
	var resp types.ProjectInfo
	err := m.client.Call("Plugin.GetProjectInfo", project, &resp)
	return &resp, err
}

func (m *AgentServiceRPCClient) LivenessProbe() (*types.Empty, error) {
	err := m.client.Call("Plugin.LivenessProbe", new(interface{}), nil)
	return new(types.Empty), err
}

func (m *AgentServiceRPCClient) LivenessProbeTimeout() uint32 {
	var resp uint32
	m.client.Call("Plugin.LivenessProbeTimeout", new(interface{}), &resp)
	return resp
}
