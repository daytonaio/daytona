// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models/enums"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) CreateBackup(ctx context.Context, containerId string, backupDto dto.CreateBackupDTO) error {
	d.cache.SetBackupState(ctx, containerId, enums.BackupStatePending)

	log.Infof("Creating backup for container %s...", containerId)

	d.cache.SetBackupState(ctx, containerId, enums.BackupStateInProgress)

	err := d.commitContainer(ctx, containerId, backupDto.Snapshot)
	if err != nil {
		return err
	}

	err = d.PushImage(ctx, backupDto.Snapshot, &backupDto.Registry)
	if err != nil {
		return err
	}

	d.cache.SetBackupState(ctx, containerId, enums.BackupStateCompleted)

	log.Infof("Backup (%s) for container %s created successfully", backupDto.Snapshot, containerId)

	err = d.RemoveImage(ctx, backupDto.Snapshot, true)
	if err != nil {
		log.Errorf("Error removing image %s: %v", backupDto.Snapshot, err)
	}

	return nil
}
