package plugin

import (
	"context"
	"errors"
	"os"
	"path"

	"github.com/daytonaio/daytona/agent/workspace"
	"github.com/daytonaio/daytona/plugins/provisioner/docker-provisioner/plugin/util"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type DockerProvisionerPlugin struct {
	BasePath string
}

func (p DockerProvisionerPlugin) GetName() string {
	return "docker"
}

func (p DockerProvisionerPlugin) GetVersion() string {
	return "0.0.1"
}

func (p DockerProvisionerPlugin) Configure() (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (p DockerProvisionerPlugin) SetConfig(config interface{}) error {
	return errors.New("not implemented")
}

func (p DockerProvisionerPlugin) getProjectPath(project workspace.Project) string {
	return path.Join(p.BasePath, "workspaces", project.Workspace.Name, "projects", project.Name)
}

func (p DockerProvisionerPlugin) CreateWorkspace(workspace workspace.Workspace) error {
	return nil
}

func (p DockerProvisionerPlugin) StartWorkspace(workspace workspace.Workspace) error {
	return nil
}

func (p DockerProvisionerPlugin) StopWorkspace(workspace workspace.Workspace) error {
	return nil
}

func (p DockerProvisionerPlugin) DestroyWorkspace(workspace workspace.Workspace) error {
	return nil
}

func (p DockerProvisionerPlugin) GetWorkspaceMetadata(workspace workspace.Workspace) (*interface{}, error) {
	return nil, errors.New("not implemented")
}

func (p DockerProvisionerPlugin) CreateProject(project workspace.Project) error {
	log.Info("Initializing project: ", project.Name)

	clonePath := p.getProjectPath(project)

	err := os.MkdirAll(clonePath, 0755)
	if err != nil {
		return err
	}

	err = util.CloneRepository(project, clonePath)
	if err != nil {
		return err
	}

	// TODO: Project image from config
	err = util.InitContainer(project, clonePath, "daytonaio/workspace-project")
	if err != nil {
		return err
	}

	err = util.StartContainer(project)
	if err != nil {
		return err
	}

	return nil
}

func (p DockerProvisionerPlugin) StartProject(project workspace.Project) error {
	return util.StartContainer(project)
}

func (p DockerProvisionerPlugin) StopProject(project workspace.Project) error {
	return util.StopContainer(project)
}

func (p DockerProvisionerPlugin) DestroyProject(project workspace.Project) error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	err = cli.ContainerRemove(ctx, util.GetContainerName(project), types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}

	err = cli.VolumeRemove(ctx, util.GetVolumeName(project), true)
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}

	err = os.RemoveAll(p.getProjectPath(project))
	if err != nil {
		return err
	}

	return nil
}

func (p DockerProvisionerPlugin) GetProjectInfo(project workspace.Project) (*workspace.ProjectInfo, error) {
	isRunning := true
	info, err := util.GetContainerInfo(project)
	if err != nil {
		if client.IsErrNotFound(err) {
			log.Debug("Container not found, project is not running")
			isRunning = false
		} else {
			return nil, err
		}
	}

	return &workspace.ProjectInfo{
		Name:                project.Name,
		IsRunning:           isRunning,
		Created:             info.Created,
		Started:             info.State.StartedAt,
		Finished:            info.State.FinishedAt,
		ProvisionerMetadata: info.Config.Labels,
	}, nil
}
