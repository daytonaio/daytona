package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/types"
)

type ProviderRPCClient struct {
	client *rpc.Client
}

func (m *ProviderRPCClient) Initialize(req InitializeProviderRequest) (*types.Empty, error) {
	err := m.client.Call("Plugin.Initialize", &req, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) GetInfo() (ProviderInfo, error) {
	var resp ProviderInfo
	err := m.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	return resp, err
}

func (m *ProviderRPCClient) Configure() (interface{}, error) {
	var config interface{}
	err := m.client.Call("Plugin.Configure", new(interface{}), &config)

	return config, err
}

func (m *ProviderRPCClient) CreateWorkspace(workspace *types.Workspace) (*types.Empty, error) {
	err := m.client.Call("Plugin.CreateWorkspace", workspace, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) StartWorkspace(workspace *types.Workspace) (*types.Empty, error) {
	err := m.client.Call("Plugin.StartWorkspace", workspace, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) StopWorkspace(workspace *types.Workspace) (*types.Empty, error) {
	err := m.client.Call("Plugin.StopWorkspace", workspace, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) DestroyWorkspace(workspace *types.Workspace) (*types.Empty, error) {
	err := m.client.Call("Plugin.DestroyWorkspace", workspace, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) GetWorkspaceInfo(workspace *types.Workspace) (*types.WorkspaceInfo, error) {
	var response types.WorkspaceInfo
	err := m.client.Call("Plugin.GetWorkspaceInfo", workspace, &response)
	return &response, err
}

func (m *ProviderRPCClient) CreateProject(project *types.Project) (*types.Empty, error) {
	err := m.client.Call("Plugin.CreateProject", project, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) StartProject(project *types.Project) (*types.Empty, error) {
	err := m.client.Call("Plugin.StartProject", project, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) StopProject(project *types.Project) (*types.Empty, error) {
	err := m.client.Call("Plugin.StopProject", project, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) DestroyProject(project *types.Project) (*types.Empty, error) {
	err := m.client.Call("Plugin.DestroyProject", project, new(types.Empty))
	return new(types.Empty), err
}

func (m *ProviderRPCClient) GetProjectInfo(project *types.Project) (*types.ProjectInfo, error) {
	var resp types.ProjectInfo
	err := m.client.Call("Plugin.GetProjectInfo", project, &resp)
	return &resp, err
}
