// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"io"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
)

func (d *DockerClient) PushImage(imageName string, cr *models.ContainerRegistry, logWriter io.Writer) error {
	ctx := context.Background()

	if logWriter != nil {
		logWriter.Write([]byte("Pushing image...\n"))
	}
	responseBody, err := d.apiClient.ImagePush(ctx, imageName, image.PushOptions{
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
		logWriter.Write([]byte("Image pushed successfully\n"))
	}

	return nil
}
