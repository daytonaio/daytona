// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models/enums"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) CreateSnapshot(ctx context.Context, containerId string, snapshotDto dto.CreateSnapshotDTO) error {
	d.cache.SetSnapshotState(ctx, containerId, enums.SnapshotStatePending)

	log.Infof("Creating snapshot for container %s...", containerId)

	d.cache.SetSnapshotState(ctx, containerId, enums.SnapshotStateInProgress)

	err := d.commitContainer(ctx, containerId, snapshotDto.Image)
	if err != nil {
		return err
	}

	err = d.PushImage(ctx, snapshotDto.Image, &snapshotDto.Registry)
	if err != nil {
		return err
	}

	d.cache.SetSnapshotState(ctx, containerId, enums.SnapshotStateCompleted)

	log.Infof("Snapshot (%s) for container %s created successfully", snapshotDto.Image, containerId)

	err = d.RemoveImage(ctx, snapshotDto.Image, true)
	if err != nil {
		log.Errorf("Error removing image %s: %v", snapshotDto.Image, err)
	}

	return nil
}
