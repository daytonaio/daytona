package plugin

import (
	"errors"

	"github.com/daytonaio/daytona/agent/workspace"
)

type ProvisionerPluginProfile struct {
	name   string
	config interface{}
}

type ProvisionerPlugin interface {
	GetName() string
	GetVersion() string

	//	client side profile config wizard
	Configure() (interface{}, error)

	//	WorkspacePreCreate
	CreateWorkspace(workspace workspace.Workspace) error
	//	WorkspacePostCreate
	//	WorkspacePreStart
	StartWorkspace(workspace workspace.Workspace) error
	//	WorkspacePostStart
	//	WorkspacePreStop
	StopWorkspace(workspace workspace.Workspace) error
	//	WorkspacePostStop
	//	WorkspacePreStop
	DestroyWorkspace(workspace workspace.Workspace) error
	//	WorkspacePostStop
	GetWorkspaceMetadata(workspace workspace.Workspace) (*interface{}, error)

	CreateProject(project workspace.Project) error
	StartProject(project workspace.Project) error
	StopProject(project workspace.Project) error
	DestroyProject(project workspace.Project) error

	// TODO: rethink name
	GetProjectInfo(project workspace.Project) (*workspace.ProjectInfo, error)
}

var provisionerPlugins []ProvisionerPlugin = []ProvisionerPlugin{}

func GetProvisionerPlugin(name string) (*ProvisionerPlugin, error) {
	//	todo
	return nil, errors.New("not implemented")
}

func GetProvisionerPlugins() []ProvisionerPlugin {
	return provisionerPlugins
}

func RegisterProvisionerPlugin(plugin ProvisionerPlugin) {
	provisionerPlugins = append(provisionerPlugins, plugin)
}
