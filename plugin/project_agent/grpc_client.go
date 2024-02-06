package project_agent

import (
	"context"

	"github.com/daytonaio/daytona/grpc/proto/types"
	"github.com/daytonaio/daytona/plugin/project_agent/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"
)

type ProjectAgentGrpcClient struct{ client proto.ProjectAgentClient }

func (m *ProjectAgentGrpcClient) GetName() (string, error) {
	resp, err := m.client.GetName(context.Background(), &empty.Empty{})
	if err != nil {
		return "", err
	}
	return resp.Name, nil
}

func (m *ProjectAgentGrpcClient) GetVersion() (string, error) {
	resp, err := m.client.GetVersion(context.Background(), &empty.Empty{})
	if err != nil {
		return "", err
	}
	return resp.Version, nil
}

func (m *ProjectAgentGrpcClient) SetConfig(config *proto.ProjectAgentConfig) error {
	_, err := m.client.SetConfig(context.Background(), config)
	return err
}

func (m *ProjectAgentGrpcClient) ProjectPreInit(project *types.Project) error {
	_, err := m.client.ProjectPreInit(context.Background(), project)
	return err
}

func (m *ProjectAgentGrpcClient) ProjectPostInit(project *types.Project) error {
	_, err := m.client.ProjectPostInit(context.Background(), project)
	return err
}

func (m *ProjectAgentGrpcClient) ProjectPreStart(project *types.Project) error {
	_, err := m.client.ProjectPreStart(context.Background(), project)
	return err
}

func (m *ProjectAgentGrpcClient) ProjectPostStart(project *types.Project) error {
	_, err := m.client.ProjectPostStart(context.Background(), project)
	return err
}

func (m *ProjectAgentGrpcClient) ProjectPreStop(project *types.Project) error {
	_, err := m.client.ProjectPreStop(context.Background(), project)
	return err
}

func (m *ProjectAgentGrpcClient) GetProjectInfo(project *types.Project) (*types.ProjectInfo, error) {
	return m.client.GetProjectInfo(context.Background(), project)
}

func (m *ProjectAgentGrpcClient) LivenessProbe() error {
	_, err := m.client.LivenessProbe(context.Background(), &empty.Empty{})
	return err
}

func (m *ProjectAgentGrpcClient) LivenessProbeTimeout() uint32 {
	resp, err := m.client.LivenessProbeTimeout(context.Background(), &empty.Empty{})
	if err != nil {
		return 0
	}
	return resp.Timeout
}
