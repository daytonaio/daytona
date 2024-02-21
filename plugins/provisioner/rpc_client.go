package provisioner

import (
	"net/rpc"

	"github.com/daytonaio/daytona/common/types"
)

type ProvisionerRPCClient struct {
	client *rpc.Client
}

func (m *ProvisionerRPCClient) Initialize(req InitializeProvisionerRequest) (Empty, error) {
	err := m.client.Call("Plugin.Initialize", req, new(interface{}))
	return Empty{}, err
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

func (m *ProvisionerRPCClient) CreateWorkspace(workspace *types.Workspace) error {
	return m.client.Call("Plugin.CreateWorkspace", workspace, new(interface{}))
}

func (m *ProvisionerRPCClient) StartWorkspace(workspace *types.Workspace) error {
	return m.client.Call("Plugin.StartWorkspace", workspace, new(interface{}))
}

func (m *ProvisionerRPCClient) StopWorkspace(workspace *types.Workspace) error {
	return m.client.Call("Plugin.StopWorkspace", workspace, new(interface{}))
}

func (m *ProvisionerRPCClient) DestroyWorkspace(workspace *types.Workspace) error {
	return m.client.Call("Plugin.DestroyWorkspace", workspace, new(interface{}))
}

func (m *ProvisionerRPCClient) GetWorkspaceInfo(workspace *types.Workspace) (*types.WorkspaceInfo, error) {
	var response types.WorkspaceInfo
	err := m.client.Call("Plugin.GetWorkspaceInfo", workspace, &response)
	return &response, err
}

func (m *ProvisionerRPCClient) CreateProject(project *types.Project) error {
	return m.client.Call("Plugin.CreateProject", project, new(interface{}))
}

func (m *ProvisionerRPCClient) StartProject(project *types.Project) error {
	return m.client.Call("Plugin.StartProject", project, new(interface{}))
}

func (m *ProvisionerRPCClient) StopProject(project *types.Project) error {
	return m.client.Call("Plugin.StopProject", project, new(interface{}))
}

func (m *ProvisionerRPCClient) DestroyProject(project *types.Project) error {
	return m.client.Call("Plugin.DestroyProject", project, new(interface{}))
}

func (m *ProvisionerRPCClient) GetProjectInfo(project *types.Project) (*types.ProjectInfo, error) {
	var resp types.ProjectInfo
	err := m.client.Call("Plugin.GetProjectInfo", project, &resp)
	return &resp, err
}
