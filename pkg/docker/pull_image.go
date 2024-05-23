// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
)

func (d *DockerClient) PullImage(imageName string, cr *containerregistry.ContainerRegistry, logWriter io.Writer) error {
	ctx := context.Background()

	tag := "latest"
	tagSplit := strings.Split(imageName, ":")
	if len(tagSplit) == 2 {
		tag = tagSplit[1]
	}

	if tag != "latest" {
		images, err := d.apiClient.ImageList(ctx, image.ListOptions{
			Filters: filters.NewArgs(filters.Arg("reference", imageName)),
		})
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

		if found {
			if logWriter != nil {
				logWriter.Write([]byte("Image already pulled\n"))
			}
			return nil
		}
	}

	if logWriter != nil {
		logWriter.Write([]byte("Pulling image...\n"))
	}
	responseBody, err := d.apiClient.ImagePull(ctx, imageName, image.PullOptions{
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
		// Sometimes registry auth fails if "" is sent, so sending "empty" instead
		return "empty"
	}

	authConfig := registry.AuthConfig{
		Username: cr.Username,
		Password: cr.Password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		// Sometimes registry auth fails if "" is sent, so sending "empty" instead
		return "empty"
	}

	return base64.URLEncoding.EncodeToString(encodedJSON)
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
