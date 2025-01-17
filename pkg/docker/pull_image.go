// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/jsonmessage"
)

func (d *DockerClient) PullImage(imageName string, cr *models.ContainerRegistry, logWriter io.Writer) error {
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

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, logWriter, 0, true, nil)
	if err != nil {
		return err
	}
	if logWriter != nil {
		logWriter.Write([]byte(views.GetPrettyLogLine("Image pulled successfully")))
	}

	return nil
}

func getRegistryAuth(cr *models.ContainerRegistry) string {
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
