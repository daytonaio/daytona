package agent_service

import (
	"context"

	"github.com/daytonaio/daytona/common/grpc/proto/types"
	"github.com/daytonaio/daytona/plugins/agent_service/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

type AgentServiceGrpcServer struct {
	Impl AgentService
}

func (m *AgentServiceGrpcServer) Initialize(ctx context.Context, req *proto.InitializeAgentServiceRequest) (*empty.Empty, error) {
	err := m.Impl.Initialize(req)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *AgentServiceGrpcServer) GetInfo(ctx context.Context, req *empty.Empty) (*proto.AgentServiceInfo, error) {
	return m.Impl.GetInfo()
}

func (m *AgentServiceGrpcServer) SetConfig(ctx context.Context, config *proto.AgentServiceConfig) (*empty.Empty, error) {
	err := m.Impl.SetConfig(config)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *AgentServiceGrpcServer) ProjectPreInit(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPreInit(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *AgentServiceGrpcServer) ProjectPostInit(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPostInit(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *AgentServiceGrpcServer) ProjectPreStart(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPreStart(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *AgentServiceGrpcServer) ProjectPostStart(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPostStart(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *AgentServiceGrpcServer) ProjectPreStop(ctx context.Context, project *types.Project) (*empty.Empty, error) {
	err := m.Impl.ProjectPreStop(project)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *AgentServiceGrpcServer) GetProjectInfo(ctx context.Context, project *types.Project) (*types.ProjectInfo, error) {
	return m.Impl.GetProjectInfo(project)
}

func (m *AgentServiceGrpcServer) LivenessProbe(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	err := m.Impl.LivenessProbe()
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (m *AgentServiceGrpcServer) LivenessProbeTimeout(ctx context.Context, _ *empty.Empty) (*proto.LivenessProbeTimeoutResponse, error) {
	timeout := m.Impl.LivenessProbeTimeout()

	return &proto.LivenessProbeTimeoutResponse{
		Timeout: timeout,
	}, nil
}
