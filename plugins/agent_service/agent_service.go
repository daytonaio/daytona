package agent_service

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/daytonaio/daytona/common/grpc/proto/types"
	"github.com/daytonaio/daytona/plugins/agent_service/grpc/proto"
)

type AgentService interface {
	Initialize(*proto.InitializeAgentServiceRequest) error
	GetInfo() (*proto.AgentServiceInfo, error)
	SetConfig(config *proto.AgentServiceConfig) error
	ProjectPreInit(project *types.Project) error
	ProjectPostInit(project *types.Project) error
	ProjectPreStart(project *types.Project) error
	ProjectPostStart(project *types.Project) error
	ProjectPreStop(project *types.Project) error
	GetProjectInfo(project *types.Project) (*types.ProjectInfo, error)
	LivenessProbe() error
	LivenessProbeTimeout() uint32
}

type AgentServicePlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl AgentService
}

func (p *AgentServicePlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterAgentServiceServer(s, &AgentServiceGrpcServer{Impl: p.Impl})
	return nil
}

func (p *AgentServicePlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &AgentServiceGrpcClient{client: proto.NewAgentServiceClient(c)}, nil
}
