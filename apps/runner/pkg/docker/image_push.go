// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"io"

	"github.com/daytonaio/runner/internal/util"
	"github.com/daytonaio/runner/pkg/api/dto"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) PushImage(ctx context.Context, imageName string, reg *dto.RegistryDTO) error {
	log.Infof("Pushing image %s...", imageName)

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

	log.Infof("Image %s pushed successfully", imageName)

	return nil
}
