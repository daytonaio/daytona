package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/grpc/proto/types"
	"github.com/daytonaio/daytona/grpc/utils"
	"github.com/daytonaio/daytona/plugin/provisioner/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

type ProvisionerGrpcServer struct {
	proto.UnimplementedProvisionerServer
	Impl Provisioner
}

func (m *ProvisionerGrpcServer) GetName(ctx context.Context, req *empty.Empty) (*proto.GetNameResponse, error) {
	name := m.Impl.GetName()
	return &proto.GetNameResponse{Name: name}, nil
}

func (m *ProvisionerGrpcServer) GetVersion(ctx context.Context, req *empty.Empty) (*proto.GetVersionResponse, error) {
	version := m.Impl.GetVersion()
	return &proto.GetVersionResponse{Version: version}, nil
}

func (m *ProvisionerGrpcServer) Configure(ctx context.Context, req *empty.Empty) (*proto.ConfigureResponse, error) {
	config, err := m.Impl.Configure()
	if err != nil {
		return nil, err
	}

	protobufConfig, err := utils.StructToProtobufStruct(config)
	if err != nil {
		return nil, err
	}

	return &proto.ConfigureResponse{Config: protobufConfig}, nil
}

func (m *ProvisionerGrpcServer) CreateWorkspace(ctx context.Context, workspace *types.Workspace) (*empty.Empty, error) {
	err := m.Impl.CreateWorkspace(workspace)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProvisionerGrpcServer) StartWorkspace(ctx context.Context, workspace *types.Workspace) (*empty.Empty, error) {
	err := m.Impl.StartWorkspace(workspace)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProvisionerGrpcServer) StopWorkspace(ctx context.Context, workspace *types.Workspace) (*empty.Empty, error) {
	err := m.Impl.StopWorkspace(workspace)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProvisionerGrpcServer) DestroyWorkspace(ctx context.Context, workspace *types.Workspace) (*empty.Empty, error) {
	err := m.Impl.DestroyWorkspace(workspace)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProvisionerGrpcServer) GetWorkspaceInfo(ctx context.Context, workspace *types.Workspace) (*types.WorkspaceInfo, error) {
	return m.Impl.GetWorkspaceInfo(workspace)
}

func (m *ProvisionerGrpcServer) CreateProject(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.CreateProject(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProvisionerGrpcServer) StartProject(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.StartProject(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProvisionerGrpcServer) StopProject(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.StopProject(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProvisionerGrpcServer) DestroyProject(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.DestroyProject(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProvisionerGrpcServer) GetProjectInfo(ctx context.Context, project *types.Project) (*types.ProjectInfo, error) {
	return m.Impl.GetProjectInfo(project)
}
