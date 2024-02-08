package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/grpc/proto/types"
	"github.com/daytonaio/daytona/plugin/provisioner/grpc/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type ProvisionerProfile struct {
	Name   string
	Config interface{}
}

type Provisioner interface {
	Initialize(*proto.InitializeProvisionerRequest) error
	GetInfo() (*proto.ProvisionerInfo, error)

	//	client side profile config wizard
	Configure() (interface{}, error)

	//	WorkspacePreCreate
	CreateWorkspace(workspace *types.Workspace) error
	//	WorkspacePostCreate
	//	WorkspacePreStart
	StartWorkspace(workspace *types.Workspace) error
	//	WorkspacePostStart
	//	WorkspacePreStop
	StopWorkspace(workspace *types.Workspace) error
	//	WorkspacePostStop
	//	WorkspacePreStop
	DestroyWorkspace(workspace *types.Workspace) error
	//	WorkspacePostStop
	GetWorkspaceInfo(workspace *types.Workspace) (*types.WorkspaceInfo, error)

	CreateProject(project *types.Project) error
	StartProject(project *types.Project) error
	StopProject(project *types.Project) error
	DestroyProject(project *types.Project) error

	// TODO: rethink name
	GetProjectInfo(project *types.Project) (*types.ProjectInfo, error)
}

type ProvisionerPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl Provisioner
}

func (p *ProvisionerPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterProvisionerServer(s, &ProvisionerGrpcServer{Impl: p.Impl})
	return nil
}

func (p *ProvisionerPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &ProvisionerGrpcClient{client: proto.NewProvisionerClient(c)}, nil
}
