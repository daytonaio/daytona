// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"strings"
)

func (d *DockerClient) TagImage(ctx context.Context, sourceImage string, targetImage string) error {
	if d.logWriter != nil {
		d.logWriter.Write([]byte(fmt.Sprintf("Tagging image %s as %s...\n", sourceImage, targetImage)))
	}

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

	if d.logWriter != nil {
		d.logWriter.Write([]byte(fmt.Sprintf("Image tagged successfully: %s → %s\n", sourceImage, targetImage)))
	}

	return nil
}
