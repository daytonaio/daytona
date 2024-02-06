package util

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/daytonaio/daytona/agent/workspace"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	log "github.com/sirupsen/logrus"
)

func InitContainer(project workspace.Project, workdirPath string, imageName string) error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return err
	}

	found := false
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if strings.HasPrefix(tag, imageName) {
				found = true
				break
			}
		}
	}

	if !found {
		log.Info("Image not found, pulling...")
		responseBody, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		defer responseBody.Close()
		_, err = io.Copy(io.Discard, responseBody)
		if err != nil {
			return err
		}
		log.Info("Image pulled successfully")
	}

	mounts := []mount.Mount{
		{
			Type:   mount.TypeVolume,
			Source: GetVolumeName(project),
			Target: "/var/lib/docker",
		},
	}

	envVars := []string{
		"DAYTONA_WS_NAME=" + project.Workspace.Name,
		"DAYTONA_WS_DIR=" + project.Name,
		"DAYTONA_WS_PROJECT_NAME=" + project.Name,
		"DAYTONA_WS_PROJECT_REPOSITORY_URL=" + project.Repository.Url,
	}

	_, err = cli.ContainerCreate(ctx, &container.Config{
		Hostname: project.Name,
		Image:    imageName,
		Labels: map[string]string{
			"daytona.workspace.name":                   project.Workspace.Name,
			"daytona.workspace.project.name":           project.Name,
			"daytona.workspace.project.repository.url": project.Repository.Url,
			// todo: Add more properties here
		},
		Env: envVars,
	}, &container.HostConfig{
		Privileged: true,
		Binds: []string{
			fmt.Sprintf("%s:/%s", workdirPath, project.Name),
			// project.GetSetupPath() + ":/setup",
			"/tmp/daytona:/tmp/daytona",
		},
		Mounts:      mounts,
		NetworkMode: container.NetworkMode(project.Workspace.Name),
	}, nil, nil, GetContainerName(project)) //	TODO: namespaced names
	if err != nil {
		return err
	}

	return nil
}
