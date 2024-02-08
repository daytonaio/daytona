package agent_service

import (
	"context"

	"github.com/daytonaio/daytona/common/grpc/proto/types"
	"github.com/daytonaio/daytona/plugins/agent_service/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

type AgentServiceGrpcClient struct{ client proto.AgentServiceClient }

func (m *AgentServiceGrpcClient) Initialize(req *proto.InitializeAgentServiceRequest) error {
	_, err := m.client.Initialize(context.Background(), req)
	return err
}

func (m *AgentServiceGrpcClient) GetInfo() (*proto.AgentServiceInfo, error) {
	return m.client.GetInfo(context.Background(), &empty.Empty{})
}

func (m *AgentServiceGrpcClient) SetConfig(config *proto.AgentServiceConfig) error {
	_, err := m.client.SetConfig(context.Background(), config)
	return err
}

func (m *AgentServiceGrpcClient) ProjectPreInit(project *types.Project) error {
	_, err := m.client.ProjectPreInit(context.Background(), project)
	return err
}

func (m *AgentServiceGrpcClient) ProjectPostInit(project *types.Project) error {
	_, err := m.client.ProjectPostInit(context.Background(), project)
	return err
}

func (m *AgentServiceGrpcClient) ProjectPreStart(project *types.Project) error {
	_, err := m.client.ProjectPreStart(context.Background(), project)
	return err
}

func (m *AgentServiceGrpcClient) ProjectPostStart(project *types.Project) error {
	_, err := m.client.ProjectPostStart(context.Background(), project)
	return err
}

func (m *AgentServiceGrpcClient) ProjectPreStop(project *types.Project) error {
	_, err := m.client.ProjectPreStop(context.Background(), project)
	return err
}

func (m *AgentServiceGrpcClient) GetProjectInfo(project *types.Project) (*types.ProjectInfo, error) {
	return m.client.GetProjectInfo(context.Background(), project)
}

func (m *AgentServiceGrpcClient) LivenessProbe() error {
	_, err := m.client.LivenessProbe(context.Background(), &empty.Empty{})
	return err
}

func (m *AgentServiceGrpcClient) LivenessProbeTimeout() uint32 {
	resp, err := m.client.LivenessProbeTimeout(context.Background(), &empty.Empty{})
	if err != nil {
		return 0
	}
	return resp.Timeout
}
