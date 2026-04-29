// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"io"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/runner/pkg/api/dto"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
)

func (d *DockerClient) PushImage(ctx context.Context, imageName string, reg *dto.RegistryDTO) error {
	d.logger.InfoContext(ctx, "Pushing image", "imageName", imageName)

	responseBody, err := d.apiClient.ImagePush(ctx, imageName, image.PushOptions{
		RegistryAuth: getRegistryAuth(reg),
	})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, io.Writer(&log.DebugLogWriter{}), 0, true, nil)
	if err != nil {
		return err
	}

	d.logger.InfoContext(ctx, "Image pushed successfully", "imageName", imageName)

	return nil
}
