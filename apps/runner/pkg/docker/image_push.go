// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"io"
	"log/slog"

	"github.com/daytonaio/runner/internal/util"
	"github.com/daytonaio/runner/pkg/api/dto"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
)

func (d *DockerClient) PushImage(ctx context.Context, imageName string, reg *dto.RegistryDTO) error {
	slog.InfoContext(ctx, "Pushing image", "imageName", imageName)

	responseBody, err := d.apiClient.ImagePush(ctx, imageName, image.PushOptions{
		RegistryAuth: getRegistryAuth(reg),
	})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, io.Writer(&util.DebugLogWriter{}), 0, true, nil)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Image pushed successfully", "imageName", imageName)

	return nil
}
