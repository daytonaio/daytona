// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/runner/pkg/api/dto"
)

func (d *DockerClient) PullSnapshot(ctx context.Context, req dto.PullSnapshotRequestDTO) error {
	// Pull the image using the pull registry (or none for public images)
	err := d.PullImage(ctx, req.Snapshot, req.Registry)
	if err != nil {
		return err
	}

	if req.DestinationRegistry != nil {
		if req.DestinationRegistry.Project == nil {
			return common_errors.NewBadRequestError(errors.New("project is required when pushing to registry"))
		}

		var targetRef string

		// If destination ref is provided, use it directly; otherwise build it from the image info
		if req.DestinationRef != nil {
			targetRef = *req.DestinationRef
		} else {
			// Get image info to retrieve the hash
			imageInfo, err := d.GetImageInfo(ctx, req.Snapshot)
			if err != nil {
				return err
			}

			ref := "daytona-" + getHashWithoutPrefix(imageInfo.Hash) + ":daytona"
			targetRef = fmt.Sprintf("%s/%s/%s", req.DestinationRegistry.Url, *req.DestinationRegistry.Project, ref)
		}

		// Tag the image for the target registry
		err = d.TagImage(ctx, req.Snapshot, targetRef)
		if err != nil {
			return err
		}

		// Push the tagged image
		err = d.PushImage(ctx, targetRef, req.DestinationRegistry)
		if err != nil {
			return err
		}
	}

	return nil
}

func getHashWithoutPrefix(hash string) string {
	return strings.TrimPrefix(hash, "sha256:")
}
