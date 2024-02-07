package project_agent

import (
	"context"

	"github.com/daytonaio/daytona/grpc/proto/types"
	"github.com/daytonaio/daytona/plugin/project_agent/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

type ProjectAgentGrpcServer struct {
	Impl ProjectAgent
}

func (m *ProjectAgentGrpcServer) GetName(ctx context.Context, req *empty.Empty) (*proto.GetNameResponse, error) {
	name, err := m.Impl.GetName()
	return &proto.GetNameResponse{Name: name}, err
}

func (m *ProjectAgentGrpcServer) GetVersion(ctx context.Context, req *empty.Empty) (*proto.GetVersionResponse, error) {
	version, err := m.Impl.GetVersion()
	return &proto.GetVersionResponse{Version: version}, err
}

func (m *ProjectAgentGrpcServer) SetConfig(ctx context.Context, config *proto.ProjectAgentConfig) (*empty.Empty, error) {
	err := m.Impl.SetConfig(config)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProjectAgentGrpcServer) ProjectPreInit(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPreInit(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProjectAgentGrpcServer) ProjectPostInit(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPostInit(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProjectAgentGrpcServer) ProjectPreStart(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPreStart(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProjectAgentGrpcServer) ProjectPostStart(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPostStart(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProjectAgentGrpcServer) ProjectPreStop(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPreStop(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProjectAgentGrpcServer) GetProjectInfo(ctx context.Context, project *types.Project) (*types.ProjectInfo, error) {
	return m.Impl.GetProjectInfo(project)
}

func (m *ProjectAgentGrpcServer) LivenessProbe(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	err := m.Impl.LivenessProbe()
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *ProjectAgentGrpcServer) LivenessProbeTimeout(ctx context.Context, _ *empty.Empty) (*proto.LivenessProbeTimeoutResponse, error) {
	timeout := m.Impl.LivenessProbeTimeout()

	return &proto.LivenessProbeTimeoutResponse{
		Timeout: timeout,
	}, nil
}
