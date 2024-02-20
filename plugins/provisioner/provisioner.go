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
	Initialize(InitializeProvisionerRequest) error
	GetInfo() (ProvisionerInfo, error)

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
	Impl Provisioner
}

func (p *ProvisionerPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ProvisionerRPCServer{Impl: p.Impl}, nil
}

func (p *ProvisionerPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ProvisionerRPCClient{client: c}, nil
}
