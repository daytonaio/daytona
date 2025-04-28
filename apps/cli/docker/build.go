// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func ImageExistsLocally(ctx context.Context, dockerClient *client.Client, imageName string) (bool, error) {
	images, err := dockerClient.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list images: %w", err)
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}
	return false, nil
}

func CheckAmdArchitecture(ctx context.Context, dockerClient *client.Client, imageName string) (bool, error) {
	inspect, _, err := dockerClient.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return false, fmt.Errorf("failed to inspect image: %w", err)
	}

	x64Architectures := []string{"amd64", "x86_64"}

	for _, arch := range x64Architectures {
		if inspect.Architecture == arch {
			return true, nil
		}
	}

	return false, nil
}
