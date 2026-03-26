// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"slices"

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
	inspect, err := dockerClient.ImageInspect(ctx, imageName)
	if err != nil {
		return false, fmt.Errorf("failed to inspect image: %w", err)
	}

	x64Architectures := []string{"amd64", "x86_64"}

	if slices.Contains(x64Architectures, inspect.Architecture) {
		return true, nil
	}

	return false, nil
}
