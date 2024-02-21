package provisioner

import (
	"net/rpc"

	"github.com/daytonaio/daytona/common/types"
)

type ProvisionerRPCClient struct {
	client *rpc.Client
}

func (m *ProvisionerRPCClient) Initialize(req InitializeProvisionerRequest) (types.Empty, error) {
	err := m.client.Call("Plugin.Initialize", &req, new(types.Empty))
	return types.Empty{}, err
}

func (m *ProvisionerRPCClient) GetInfo() (ProvisionerInfo, error) {
	var resp ProvisionerInfo
	err := m.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	return resp, err
}

func (m *ProvisionerRPCClient) Configure() (interface{}, error) {
	var config interface{}
	err := m.client.Call("Plugin.Configure", new(interface{}), &config)

	return config, err
}

func (m *ProvisionerRPCClient) CreateWorkspace(workspace types.Workspace) (types.Empty, error) {
	err := m.client.Call("Plugin.CreateWorkspace", &workspace, new(types.Empty))
	return types.Empty{}, err
}

func (m *ProvisionerRPCClient) StartWorkspace(workspace *types.Workspace) (types.Empty, error) {
	err := m.client.Call("Plugin.StartWorkspace", workspace, new(types.Empty))
	return types.Empty{}, err
}

func (m *ProvisionerRPCClient) StopWorkspace(workspace *types.Workspace) (types.Empty, error) {
	err := m.client.Call("Plugin.StopWorkspace", workspace, new(types.Empty))
	return types.Empty{}, err
}

func (m *ProvisionerRPCClient) DestroyWorkspace(workspace *types.Workspace) (types.Empty, error) {
	err := m.client.Call("Plugin.DestroyWorkspace", workspace, new(types.Empty))
	return types.Empty{}, err
}

func (m *ProvisionerRPCClient) GetWorkspaceInfo(workspace *types.Workspace) (*types.WorkspaceInfo, error) {
	var response types.WorkspaceInfo
	err := m.client.Call("Plugin.GetWorkspaceInfo", workspace, &response)
	return &response, err
}

func (m *ProvisionerRPCClient) CreateProject(project *types.Project) (types.Empty, error) {
	err := m.client.Call("Plugin.CreateProject", project, new(types.Empty))
	return types.Empty{}, err
}

func (m *ProvisionerRPCClient) StartProject(project *types.Project) (types.Empty, error) {
	err := m.client.Call("Plugin.StartProject", project, new(types.Empty))
	return types.Empty{}, err
}

func (m *ProvisionerRPCClient) StopProject(project *types.Project) (types.Empty, error) {
	err := m.client.Call("Plugin.StopProject", project, new(types.Empty))
	return types.Empty{}, err
}

func (m *ProvisionerRPCClient) DestroyProject(project *types.Project) (types.Empty, error) {
	err := m.client.Call("Plugin.DestroyProject", project, new(types.Empty))
	return types.Empty{}, err
}

func (m *ProvisionerRPCClient) GetProjectInfo(project *types.Project) (*types.ProjectInfo, error) {
	var resp types.ProjectInfo
	err := m.client.Call("Plugin.GetProjectInfo", project, &resp)
	return &resp, err
}
