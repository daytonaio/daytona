package provisioner

import (
	"github.com/daytonaio/daytona/pkg/types"
)

type ProvisionerRPCServer struct {
	Impl Provisioner
}

func (m *ProvisionerRPCServer) Initialize(arg InitializeProvisionerRequest, resp *types.Empty) error {
	_, err := m.Impl.Initialize(arg)
	return err
}

func (m *ProvisionerRPCServer) GetInfo(arg interface{}, resp *ProvisionerInfo) error {
	info, err := m.Impl.GetInfo()
	if err != nil {
		return err
	}

	*resp = info
	return nil
}

func (m *ProvisionerRPCServer) Configure(arg interface{}, configResponse *interface{}) error {
	config, err := m.Impl.Configure()
	if err != nil {
		return err
	}

	*configResponse = config
	return nil
}

func (m *ProvisionerRPCServer) CreateWorkspace(arg *types.Workspace, resp *types.Empty) error {
	_, err := m.Impl.CreateWorkspace(arg)
	return err
}

func (m *ProvisionerRPCServer) StartWorkspace(arg *types.Workspace, resp *types.Empty) error {
	_, err := m.Impl.StartWorkspace(arg)
	return err
}

func (m *ProvisionerRPCServer) StopWorkspace(arg *types.Workspace, resp *types.Empty) error {
	_, err := m.Impl.StopWorkspace(arg)
	return err
}

func (m *ProvisionerRPCServer) DestroyWorkspace(arg *types.Workspace, resp *types.Empty) error {
	_, err := m.Impl.DestroyWorkspace(arg)
	return err
}

func (m *ProvisionerRPCServer) GetWorkspaceInfo(arg *types.Workspace, resp *types.WorkspaceInfo) error {
	info, err := m.Impl.GetWorkspaceInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}

func (m *ProvisionerRPCServer) CreateProject(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.CreateProject(arg)
	return err
}

func (m *ProvisionerRPCServer) StartProject(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.StartProject(arg)
	return err
}

func (m *ProvisionerRPCServer) StopProject(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.StopProject(arg)
	return err
}

func (m *ProvisionerRPCServer) DestroyProject(arg *types.Project, resp *types.Empty) error {
	_, err := m.Impl.DestroyProject(arg)
	return err
}

func (m *ProvisionerRPCServer) GetProjectInfo(arg *types.Project, resp *types.ProjectInfo) error {
	info, err := m.Impl.GetProjectInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}
