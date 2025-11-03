// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"log/slog"
	"strings"

	"github.com/docker/docker/api/types/image"
)

func (d *DockerClient) ImageExists(ctx context.Context, imageName string, includeLatest bool) (bool, error) {
	imageName = strings.Replace(imageName, "docker.io/", "", 1)

	if strings.HasSuffix(imageName, ":latest") && !includeLatest {
		return false, nil
	}

	images, err := d.apiClient.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
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
		slog.InfoContext(ctx, "Image already pulled", "imageName", imageName)
	}

	return found, nil
}
