// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/image"
)

func (d *DockerClient) DeleteImage(imageName string, force bool, logWriter io.Writer) error {
	ctx := context.Background()

	_, err := d.apiClient.ImageRemove(ctx, imageName, image.RemoveOptions{
		Force: force,
	})
	if err != nil {
		return err
	}

	if logWriter != nil {
		logWriter.Write([]byte("Image deleted successfully\n"))
	}

	return nil
}
