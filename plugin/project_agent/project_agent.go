package project_agent

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/daytonaio/daytona/grpc/proto/types"
	"github.com/daytonaio/daytona/plugin/project_agent/grpc/proto"
)

type ProjectAgent interface {
	GetName() (string, error)
	GetVersion() (string, error)
	SetConfig(config *proto.ProjectAgentConfig) error
	ProjectPreInit(project *types.Project) error
	ProjectPostInit(project *types.Project) error
	ProjectPreStart(project *types.Project) error
	ProjectPostStart(project *types.Project) error
	ProjectPreStop(project *types.Project) error
	GetProjectInfo(project *types.Project) (*types.ProjectInfo, error)
	LivenessProbe() error
	LivenessProbeTimeout() uint32
}

type ProjectAgentPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl ProjectAgent
}

func (p *ProjectAgentPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterProjectAgentServer(s, &ProjectAgentGrpcServer{Impl: p.Impl})
	return nil
}

func (p *ProjectAgentPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &ProjectAgentGrpcClient{client: proto.NewProjectAgentClient(c)}, nil
}
