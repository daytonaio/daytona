// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"log/slog"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models/enums"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type backupContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var backup_context_map = cmap.New[backupContext]()

func (d *DockerClient) StartBackupCreate(ctx context.Context, containerId string, backupDto dto.CreateBackupDTO) error {
	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	slog.InfoContext(ctx, "Creating backup for container", "containerId", containerId)

	d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateInProgress, nil)

	go func() {
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
				slog.InfoContext(ctx, "Backup for container canceled", "containerId", containerId)
				return
			}
			slog.ErrorContext(ctx, "Error committing container", "containerId", containerId, "error", err)
			d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateFailed, err)
			return
		}

		err = d.PushImage(ctx, backupDto.Snapshot, &backupDto.Registry)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateNone, nil)
				slog.InfoContext(ctx, "Backup for container canceled", "containerId", containerId)
				return
			}
			slog.ErrorContext(ctx, "Error pushing image", "snapshot", backupDto.Snapshot, "error", err)
			d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateFailed, err)
			return
		}

		d.statesCache.SetBackupState(ctx, containerId, enums.BackupStateCompleted, nil)

		slog.InfoContext(ctx, "Backup created successfully", "snapshot", backupDto.Snapshot, "containerId", containerId)

		err = d.RemoveImage(ctx, backupDto.Snapshot, true)
		if err != nil {
			slog.ErrorContext(ctx, "Error removing image", "snapshot", backupDto.Snapshot, "error", err)
			// Don't set backup to failed because the image is already pushed
		}
	}()

	return nil
}
