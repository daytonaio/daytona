package provider

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/types"
	"github.com/hashicorp/go-plugin"
)

type ProviderProfile struct {
	Name   string
	Config interface{}
}

type ProviderInfo struct {
	Name    string
	Version string
}

type InitializeProviderRequest struct {
	BasePath          string
	ServerDownloadUrl string
	ServerVersion     string
	ServerUrl         string
	ServerApiUrl      string
}

type Provider interface {
	Initialize(InitializeProviderRequest) (*types.Empty, error)
	GetInfo() (ProviderInfo, error)

	//	client side profile config wizard
	Configure() (interface{}, error)

	//	WorkspacePreCreate
	CreateWorkspace(workspace *types.Workspace) (*types.Empty, error)
	//	WorkspacePostCreate
	//	WorkspacePreStart
	StartWorkspace(workspace *types.Workspace) (*types.Empty, error)
	//	WorkspacePostStart
	//	WorkspacePreStop
	StopWorkspace(workspace *types.Workspace) (*types.Empty, error)
	//	WorkspacePostStop
	//	WorkspacePreStop
	DestroyWorkspace(workspace *types.Workspace) (*types.Empty, error)
	//	WorkspacePostStop
	GetWorkspaceInfo(workspace *types.Workspace) (*types.WorkspaceInfo, error)

	CreateProject(project *types.Project) (*types.Empty, error)
	StartProject(project *types.Project) (*types.Empty, error)
	StopProject(project *types.Project) (*types.Empty, error)
	DestroyProject(project *types.Project) (*types.Empty, error)

	// TODO: rethink name
	GetProjectInfo(project *types.Project) (*types.ProjectInfo, error)
}

type ProviderPlugin struct {
	Impl Provider
}

func (p *ProviderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ProviderRPCServer{Impl: p.Impl}, nil
}

func (p *ProviderPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ProviderRPCClient{client: c}, nil
}
