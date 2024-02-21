package provisioner

import (
	"net/rpc"

	"github.com/daytonaio/daytona/common/types"
	"github.com/hashicorp/go-plugin"
)

type ProvisionerProfile struct {
	Name   string
	Config interface{}
}

type ProvisionerInfo struct {
	Name    string
	Version string
}

type InitializeProvisionerRequest struct {
	BasePath          string
	ServerDownloadUrl string
	ServerVersion     string
	ServerUrl         string
	ServerApiUrl      string
}

type Provisioner interface {
	Initialize(InitializeProvisionerRequest) (types.Empty, error)
	GetInfo() (ProvisionerInfo, error)

	//	client side profile config wizard
	Configure() (interface{}, error)

	//	WorkspacePreCreate
	CreateWorkspace(workspace *types.Workspace) (*types.Empty, error)
	//	WorkspacePostCreate
	//	WorkspacePreStart
	StartWorkspace(workspace *types.Workspace) (types.Empty, error)
	//	WorkspacePostStart
	//	WorkspacePreStop
	StopWorkspace(workspace *types.Workspace) (types.Empty, error)
	//	WorkspacePostStop
	//	WorkspacePreStop
	DestroyWorkspace(workspace *types.Workspace) (types.Empty, error)
	//	WorkspacePostStop
	GetWorkspaceInfo(workspace *types.Workspace) (*types.WorkspaceInfo, error)

	CreateProject(project *types.Project) (types.Empty, error)
	StartProject(project *types.Project) (types.Empty, error)
	StopProject(project *types.Project) (types.Empty, error)
	DestroyProject(project *types.Project) (types.Empty, error)

	// TODO: rethink name
	GetProjectInfo(project *types.Project) (*types.ProjectInfo, error)
}

type ProvisionerPlugin struct {
	Impl Provisioner
}

func (p *ProvisionerPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ProvisionerRPCServer{Impl: p.Impl}, nil
}

func (p *ProvisionerPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ProvisionerRPCClient{client: c}, nil
}
