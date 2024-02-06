package plugin

import (
	"errors"

	"github.com/daytonaio/daytona/agent/workspace"
)

type SshProvisionerConfig struct {
	host     string
	port     int
	user     string
	password string
}

type SshProvisionerPlugin struct {
	BasePath string
}

func (e SshProvisionerPlugin) GetName() string {
	return "ssh"
}

func (e SshProvisionerPlugin) GetVersion() string {
	return "0.0.1"
}

func (e SshProvisionerPlugin) Configure() (interface{}, error) {
	//	todo: client side config wizard (charm)
	hardcoded_for_now := SshProvisionerConfig{}
	hardcoded_for_now.host = "1.1.1.1"
	hardcoded_for_now.port = 22
	hardcoded_for_now.user = "root"
	hardcoded_for_now.password = "test1234"

	return hardcoded_for_now, nil
}

func (e SshProvisionerPlugin) CreateWorkspace(workspace workspace.Workspace) error {
	return errors.New("not implemented")
}

func (e SshProvisionerPlugin) StartWorkspace(workspace workspace.Workspace) error {
	return errors.New("not implemented")
}

func (e SshProvisionerPlugin) StopWorkspace(workspace workspace.Workspace) error {
	return errors.New("not implemented")
}

func (e SshProvisionerPlugin) DestroyWorkspace(workspace workspace.Workspace) error {
	return errors.New("not implemented")
}

func (e SshProvisionerPlugin) GetWorkspaceMetadata(workspace workspace.Workspace) (*interface{}, error) {
	return nil, errors.New("not implemented")
}

func (e SshProvisionerPlugin) SetConfig(config interface{}) error {
	return errors.New("not implemented")
}

func (e SshProvisionerPlugin) CreateProject(project workspace.Project) error {
	return errors.New("not implemented")
}

func (e SshProvisionerPlugin) StartProject(project workspace.Project) error {
	return errors.New("not implemented")
}

func (e SshProvisionerPlugin) StopProject(project workspace.Project) error {
	return errors.New("not implemented")
}

func (e SshProvisionerPlugin) DestroyProject(project workspace.Project) error {
	return errors.New("not implemented")
}

func (e SshProvisionerPlugin) GetProjectInfo(project workspace.Project) (*workspace.ProjectInfo, error) {
	return nil, errors.New("not implemented")
}
