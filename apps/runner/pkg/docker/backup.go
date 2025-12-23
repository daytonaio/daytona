// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models/enums"

	cmap "github.com/orcaman/concurrent-map/v2"

	log "github.com/sirupsen/logrus"
)

type backupContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var backup_context_map = cmap.New[backupContext]()

func (d *DockerClient) CreateBackup(ctx context.Context, containerId string, backupDto dto.CreateBackupDTO) error {
	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	log.Infof("Creating backup for container %s...", containerId)

	return d.createBackup(containerId, backupDto)
}

func (d *DockerClient) CreateBackupAsync(ctx context.Context, containerId string, backupDto dto.CreateBackupDTO) error {
	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	log.Infof("Creating backup for container %s...", containerId)

	go func() {
		d.createBackup(containerId, backupDto)
	}()

	return nil
}

func (d *DockerClient) createBackup(containerId string, backupDto dto.CreateBackupDTO) error {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		backupContext, ok := backup_context_map.Get(containerId)
		if ok {
			backupContext.cancel()
		}
		backup_context_map.Remove(containerId)
	}()

	backup_context_map.Set(containerId, backupContext{ctx, cancel})

	err := d.commitContainer(ctx, containerId, backupDto.Snapshot)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateNone, nil)
			log.Infof("Backup for container %s canceled", containerId)
			return err
		}
		log.Errorf("Error committing container %s: %v", containerId, err)
		d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateFailed, err)
		return err
	}

	err = d.PushImage(ctx, backupDto.Snapshot, &backupDto.Registry)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateNone, nil)
			log.Infof("Backup for container %s canceled", containerId)
			return err
		}
		log.Errorf("Error pushing image %s: %v", backupDto.Snapshot, err)
		d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateFailed, err)
		return err
	}

	d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateCompleted, nil)

	log.Infof("Backup (%s) for container %s created successfully", backupDto.Snapshot, containerId)

	err = d.RemoveImage(ctx, backupDto.Snapshot, true)
	if err != nil {
		log.Errorf("Error removing image %s: %v", backupDto.Snapshot, err)
		// Don't set backup to failed because the image is already pushed
	}

	return nil
}
