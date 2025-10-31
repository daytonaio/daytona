// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

func (d *DockerClient) TagImage(ctx context.Context, sourceImage string, targetImage string) error {
	slog.InfoContext(ctx, "Tagging image", "sourceImage", sourceImage, "targetImage", targetImage)

	// Extract repository and tag from targetImage
	lastColonIndex := strings.LastIndex(targetImage, ":")
	var repo, tag string

	if lastColonIndex == -1 {
		return fmt.Errorf("invalid target image format: %s", targetImage)
	} else {
		repo = targetImage[:lastColonIndex]
		tag = targetImage[lastColonIndex+1:]
	}

	if repo == "" || tag == "" {
		return fmt.Errorf("invalid target image format: %s", targetImage)
	}

	err := d.apiClient.ImageTag(ctx, sourceImage, targetImage)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Image tagged successfully", "sourceImage", sourceImage, "targetImage", targetImage)

	return nil
}
