// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
)

func (d *DockerClient) CreateWorkspace(workspace *workspace.Workspace, logWriter io.Writer) error {
	if logWriter != nil {
		logWriter.Write([]byte("Initializing network\n"))
	}
	ctx := context.Background()

	networks, err := d.apiClient.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, network := range networks {
		if network.Name == workspace.Id {
			if logWriter != nil {
				logWriter.Write([]byte("Network already exists\n"))
			}
			return nil
		}
	}

	_, err = d.apiClient.NetworkCreate(ctx, workspace.Id, types.NetworkCreate{
		Attachable: true,
	})
	if err != nil {
		return err
	}

	if logWriter != nil {
		logWriter.Write([]byte("Network initialized\n"))
	}
	return nil
}

func (d *DockerClient) CreateProject(project *workspace.Project, daytonaDownloadUrl string, cr *containerregistry.ContainerRegistry, logWriter io.Writer) error {
	err := d.pullProjectImage(project, cr, logWriter)
	if err != nil {
		return err
	}

	return d.initProjectContainer(project, daytonaDownloadUrl)
}

func (d *DockerClient) pullProjectImage(project *workspace.Project, cr *containerregistry.ContainerRegistry, logWriter io.Writer) error {
	ctx := context.Background()

	tag := "latest"
	tagSplit := strings.Split(project.Image, ":")
	if len(tagSplit) == 2 {
		tag = tagSplit[1]
	}

	if tag != "latest" {
		images, err := d.apiClient.ImageList(ctx, image.ListOptions{
			Filters: filters.NewArgs(filters.Arg("reference", project.Image)),
		})
		if err != nil {
			return err
		}

		found := false
		for _, image := range images {
			for _, tag := range image.RepoTags {
				if strings.HasPrefix(tag, project.Image) {
					found = true
					break
				}
			}
		}

		if found {
			if logWriter != nil {
				logWriter.Write([]byte("Image found locally\n"))
			}
			return nil
		}
	}

	if logWriter != nil {
		logWriter.Write([]byte("Pulling image...\n"))
	}
	responseBody, err := d.apiClient.ImagePull(ctx, project.Image, image.PullOptions{
		RegistryAuth: getRegistryAuth(cr),
	})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	err = readPullProgress(responseBody, logWriter)
	if err != nil {
		return err
	}

	if logWriter != nil {
		logWriter.Write([]byte("Image pulled successfully\n"))
	}

	return nil
}

func getRegistryAuth(cr *containerregistry.ContainerRegistry) string {
	if cr == nil {
		return ""
	}

	authConfig := registry.AuthConfig{
		Username: cr.Username,
		Password: cr.Password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(encodedJSON)
}

func (d *DockerClient) initProjectContainer(project *workspace.Project, daytonaDownloadUrl string) error {
	ctx := context.Background()

	_, err := d.apiClient.ContainerCreate(ctx, GetContainerCreateConfig(project, daytonaDownloadUrl), &container.HostConfig{
		Privileged:  true,
		NetworkMode: container.NetworkMode(project.WorkspaceId),
	}, nil, nil, d.GetProjectContainerName(project))
	if err != nil {
		return err
	}

	return nil
}

func readPullProgress(pullResponse io.ReadCloser, logWriter io.Writer) error {
	if logWriter == nil {
		return nil
	}

	cursor := Cursor{
		logWriter: logWriter,
	}
	layers := make([]string, 0)
	oldIndex := len(layers)

	var event *pullEvent
	decoder := json.NewDecoder(pullResponse)

	for {
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		imageID := event.ID

		// Check if the line is one of the final two ones
		if strings.HasPrefix(event.Status, "Digest:") || strings.HasPrefix(event.Status, "Status:") {
			logWriter.Write([]byte(fmt.Sprintf("%s\n", event.Status)))
			continue
		}

		// Check if ID has already passed once
		index := 0
		for i, v := range layers {
			if v == imageID {
				index = i + 1
				break
			}
		}

		if index > 0 {
			diff := index - oldIndex

			if diff > 1 {
				down := diff - 1
				cursor.moveDown(down)
			} else if diff < 1 {
				up := diff*(-1) + 1
				cursor.moveUp(up)
			}

			oldIndex = index
		} else {
			layers = append(layers, event.ID)
			diff := len(layers) - oldIndex

			if diff > 1 {
				cursor.moveDown(diff) // Return to the last row
			}

			oldIndex = len(layers)
		}

		if event.Status == "Pull complete" {
			logWriter.Write([]byte(fmt.Sprintf("%s: %s\n", event.ID, event.Status)))
		} else {
			logWriter.Write([]byte(fmt.Sprintf("%s: %s %s\n", event.ID, event.Status, event.Progress)))
		}
	}

	return nil
}

func GetContainerCreateConfig(project *workspace.Project, daytonaDownloadUrl string) *container.Config {
	envVars := []string{
		"DAYTONA_WS_DIR=" + path.Join("/workspaces", project.Name),
	}

	for key, value := range project.EnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	return &container.Config{
		Hostname: project.Name,
		Image:    project.Image,
		Labels: map[string]string{
			"daytona.workspace.id":                     project.WorkspaceId,
			"daytona.workspace.project.name":           project.Name,
			"daytona.workspace.project.repository.url": project.Repository.Url,
		},
		//	User:         project.User,
		User:         "root",
		Env:          envVars,
		Entrypoint:   []string{"bash", "-c", util.GetProjectStartScript(daytonaDownloadUrl, project.ApiKey)},
		AttachStdout: true,
		AttachStderr: true,
	}
}

// Cursor structure that implements some methods
// for manipulating command line's cursor
type Cursor struct {
	logWriter io.Writer
}

func (c *Cursor) moveUp(rows int) {
	c.logWriter.Write([]byte(fmt.Sprintf("\033[%dF", rows)))
}

func (c *Cursor) moveDown(rows int) {
	c.logWriter.Write([]byte(fmt.Sprintf("\033[%dE", rows)))
}

type pullEvent struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	Progress       string `json:"progress,omitempty"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}
