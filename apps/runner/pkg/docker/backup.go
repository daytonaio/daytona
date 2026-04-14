// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"time"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models/enums"

	cmap "github.com/orcaman/concurrent-map/v2"
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

	d.logger.InfoContext(ctx, "Creating backup for container", "containerId", containerId)

	return d.createBackup(containerId, backupDto)
}

func (d *DockerClient) CreateBackupAsync(ctx context.Context, containerId string, backupDto dto.CreateBackupDTO) error {
	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	d.logger.InfoContext(ctx, "Creating backup for container", "containerId", containerId)

	cacheErr := d.backupInfoCache.SetBackupState(ctx, containerId, enums.BackupStateInProgress, backupDto.Snapshot, nil)
	if cacheErr != nil {
		d.logger.DebugContext(ctx, "Failed to update backup info", "error", cacheErr)
	}

	go func() {
		err := d.createBackup(containerId, backupDto)
		if err != nil {
			d.logger.ErrorContext(ctx, "Error creating backup for container", "containerId", containerId, "error", err)
		}
	}()

	return nil
}

func (d *DockerClient) createBackup(containerId string, backupDto dto.CreateBackupDTO) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.backupTimeoutMin)*time.Minute)

	defer func() {
		backupContext, ok := backup_context_map.Get(containerId)
		if ok {
			backupContext.cancel()
		}
		backup_context_map.Remove(containerId)
	}()

	backup_context_map.Set(containerId, backupContext{ctx, cancel})

	cacheErr := d.backupInfoCache.SetBackupState(ctx, containerId, enums.BackupStateInProgress, backupDto.Snapshot, nil)
	if cacheErr != nil {
		d.logger.DebugContext(ctx, "Failed to update backup info", "error", cacheErr)
	}

	err := d.commitContainer(ctx, containerId, backupDto.Snapshot)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			cacheErr := d.backupInfoCache.SetBackupState(ctx, containerId, enums.BackupStateNone, backupDto.Snapshot, nil)
			if cacheErr != nil {
				d.logger.DebugContext(ctx, "Failed to update backup info", "error", cacheErr)
			}

			d.logger.InfoContext(ctx, "Backup canceled for container", "containerId", containerId)
			return err
		}

		if errors.Is(err, context.DeadlineExceeded) {
			cacheErr := d.backupInfoCache.SetBackupState(ctx, containerId, enums.BackupStateFailed, backupDto.Snapshot, err)
			if cacheErr != nil {
				d.logger.DebugContext(ctx, "Failed to update backup info", "error", cacheErr)
			}

			d.logger.ErrorContext(ctx, "Backup timed out during commit", "containerId", containerId)
			return err
		}

		d.logger.ErrorContext(ctx, "Error committing container", "containerId", containerId, "error", err)

		cacheErr := d.backupInfoCache.SetBackupState(ctx, containerId, enums.BackupStateFailed, backupDto.Snapshot, err)
		if cacheErr != nil {
			d.logger.DebugContext(ctx, "Failed to update backup info", "error", cacheErr)
		}

		return err
	}

	err = d.PushImage(ctx, backupDto.Snapshot, backupDto.GetRegistry())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			cacheErr := d.backupInfoCache.SetBackupState(ctx, containerId, enums.BackupStateNone, backupDto.Snapshot, nil)
			if cacheErr != nil {
				d.logger.DebugContext(ctx, "Failed to update backup info", "error", cacheErr)
			}

			d.logger.InfoContext(ctx, "Backup canceled for container", "containerId", containerId)
			return err
		}

		if errors.Is(err, context.DeadlineExceeded) {
			cacheErr := d.backupInfoCache.SetBackupState(ctx, containerId, enums.BackupStateFailed, backupDto.Snapshot, err)
			if cacheErr != nil {
				d.logger.DebugContext(ctx, "Failed to update backup info", "error", cacheErr)
			}

			d.logger.ErrorContext(ctx, "Backup timed out during push", "containerId", containerId)
			return err
		}

		d.logger.ErrorContext(ctx, "Error pushing image", "image", backupDto.Snapshot, "error", err)

		cacheErr := d.backupInfoCache.SetBackupState(ctx, containerId, enums.BackupStateFailed, backupDto.Snapshot, err)
		if cacheErr != nil {
			d.logger.DebugContext(ctx, "Failed to update backup info", "error", cacheErr)
		}

		return err
	}

	cacheErr = d.backupInfoCache.SetBackupState(ctx, containerId, enums.BackupStateCompleted, backupDto.Snapshot, nil)
	if cacheErr != nil {
		d.logger.DebugContext(ctx, "Failed to update backup info", "error", cacheErr)
	}

	d.logger.InfoContext(ctx, "Backup created successfully", "snapshot", backupDto.Snapshot, "containerId", containerId)

	err = d.RemoveImage(ctx, backupDto.Snapshot, true)
	if err != nil {
		d.logger.ErrorContext(ctx, "Error removing image", "image", backupDto.Snapshot, "error", err)
		// Don't set backup to failed because the image is already pushed
	}

	return nil
}
