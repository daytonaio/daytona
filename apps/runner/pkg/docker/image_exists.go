// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
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
		// Check RepoTags for tag-based references
		for _, tag := range image.RepoTags {
			if strings.HasPrefix(tag, imageName) {
				found = true
				break
			}
		}
		if found {
			break
		}

		// Check RepoDigests for digest-based references
		for _, digest := range image.RepoDigests {
			if strings.HasPrefix(digest, imageName) || strings.HasSuffix(digest, strings.TrimPrefix(imageName, "library/")) {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if found {
		d.logger.InfoContext(ctx, "Image already pulled", "imageName", imageName)
	}

	return found, nil
}
