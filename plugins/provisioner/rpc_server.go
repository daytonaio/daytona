package provisioner

import (
	"github.com/daytonaio/daytona/common/types"
)

type ProvisionerRPCServer struct {
	Impl Provisioner
}

func (m *ProvisionerRPCServer) Initialize(arg InitializeProvisionerRequest, resp *interface{}) error {
	res, err := m.Impl.Initialize(arg)
	if err != nil {
		return err
	}

	*resp = res
	return nil
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

func (m *ProvisionerRPCServer) CreateWorkspace(arg *types.Workspace, resp interface{}) error {
	return m.Impl.CreateWorkspace(arg)
}

func (m *ProvisionerRPCServer) StartWorkspace(arg *types.Workspace, resp interface{}) error {
	return m.Impl.StartWorkspace(arg)
}

func (m *ProvisionerRPCServer) StopWorkspace(arg *types.Workspace, resp interface{}) error {
	return m.Impl.StopWorkspace(arg)
}

func (m *ProvisionerRPCServer) DestroyWorkspace(arg *types.Workspace, resp interface{}) error {
	return m.Impl.DestroyWorkspace(arg)
}

func (m *ProvisionerRPCServer) GetWorkspaceInfo(arg *types.Workspace, resp *types.WorkspaceInfo) error {
	info, err := m.Impl.GetWorkspaceInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}

func (m *ProvisionerRPCServer) CreateProject(arg *types.Project, resp interface{}) error {
	return m.Impl.CreateProject(arg)
}

func (m *ProvisionerRPCServer) StartProject(arg *types.Project, resp interface{}) error {
	return m.Impl.StartProject(arg)
}

func (m *ProvisionerRPCServer) StopProject(arg *types.Project, resp interface{}) error {
	return m.Impl.StopProject(arg)
}

func (m *ProvisionerRPCServer) DestroyProject(arg *types.Project, resp interface{}) error {
	return m.Impl.DestroyProject(arg)
}

func (m *ProvisionerRPCServer) GetProjectInfo(arg *types.Project, resp *types.ProjectInfo) error {
	info, err := m.Impl.GetProjectInfo(arg)
	if err != nil {
		return err
	}

	*resp = *info
	return nil
}
