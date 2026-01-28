// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"

	"github.com/daytonaio/runner-win/pkg/storage"
	log "github.com/sirupsen/logrus"
)

// RemoveImage removes a snapshot from object storage
func (l *LibVirt) RemoveImage(ctx context.Context, imageName string, force bool) error {
	log.Infof("RemoveImage: %s (force=%v)", imageName, force)

	// Get the storage client
	storageClient, err := storage.GetObjectStorageClient()
	if err != nil {
		log.Warnf("Failed to get storage client, skipping object storage deletion: %v", err)
		return nil
	}

	// Delete from object storage
	// The imageName is the snapshot ref which includes the path in object storage
	if err := storageClient.DeleteSnapshot(ctx, imageName); err != nil {
		log.Warnf("Failed to delete snapshot from object storage: %v", err)
		// Return error only if force is false
		if !force {
			return err
		}
		// With force=true, we log the error but continue
		return nil
	}

	log.Infof("Successfully deleted snapshot '%s' from object storage", imageName)
	return nil
}
