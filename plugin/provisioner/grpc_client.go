package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/grpc/proto/types"
	"github.com/daytonaio/daytona/plugin/provisioner/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

type ProvisionerGrpcClient struct {
	client proto.ProvisionerClient
}

func (m *ProvisionerGrpcClient) GetName() (string, error) {
	resp, err := m.client.GetName(context.Background(), &empty.Empty{})
	if err != nil {
		return "", err
	}

	return resp.Name, nil
}

func (m *ProvisionerGrpcClient) GetVersion() (string, error) {
	resp, err := m.client.GetVersion(context.Background(), &empty.Empty{})
	if err != nil {
		return "", err
	}

	return resp.Version, nil
}

func (m *ProvisionerGrpcClient) Configure() (interface{}, error) {
	resp, err := m.client.Configure(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return resp.Config, nil
}

func (m *ProvisionerGrpcClient) CreateWorkspace(workspace *types.Workspace) error {
	_, err := m.client.CreateWorkspace(context.Background(), workspace)
	return err
}

func (m *ProvisionerGrpcClient) StartWorkspace(workspace *types.Workspace) error {
	_, err := m.client.StartWorkspace(context.Background(), workspace)
	return err
}

func (m *ProvisionerGrpcClient) StopWorkspace(workspace *types.Workspace) error {
	_, err := m.client.StopWorkspace(context.Background(), workspace)
	return err
}

func (m *ProvisionerGrpcClient) DestroyWorkspace(workspace *types.Workspace) error {
	_, err := m.client.DestroyWorkspace(context.Background(), workspace)
	return err
}

func (m *ProvisionerGrpcClient) GetWorkspaceInfo(workspace *types.Workspace) (*types.WorkspaceInfo, error) {
	return m.client.GetWorkspaceInfo(context.Background(), workspace)
}

func (m *ProvisionerGrpcClient) CreateProject(project *types.Project) error {
	_, err := m.client.CreateProject(context.Background(), project)
	return err
}

func (m *ProvisionerGrpcClient) StartProject(project *types.Project) error {
	_, err := m.client.StartProject(context.Background(), project)
	return err
}

func (m *ProvisionerGrpcClient) StopProject(project *types.Project) error {
	_, err := m.client.StopProject(context.Background(), project)
	return err
}

func (m *ProvisionerGrpcClient) DestroyProject(project *types.Project) error {
	_, err := m.client.DestroyProject(context.Background(), project)
	return err
}

func (m *ProvisionerGrpcClient) GetProjectInfo(project *types.Project) (*types.ProjectInfo, error) {
	return m.client.GetProjectInfo(context.Background(), project)
}
