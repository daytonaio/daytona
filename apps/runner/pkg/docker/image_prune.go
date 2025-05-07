// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/filters"
	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) PruneImages(ctx context.Context, pruneFilters filters.Args) (int, uint64, error) {
	pruneReport, err := d.apiClient.ImagesPrune(ctx, pruneFilters)
	if err != nil {
		return 0, 0, err
	}

	return len(pruneReport.ImagesDeleted), pruneReport.SpaceReclaimed, nil
}

func (d *DockerClient) PruneDanglingImages(ctx context.Context) {
	go func() {
		// Run cleanup every hour
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Info("Pruning dangling images")

				pruneFilters := filters.NewArgs()
				pruneFilters.Add("dangling", "true")

				prunedImages, spaceReclaimed, err := d.PruneImages(ctx, pruneFilters)
				if err != nil {
					log.Errorf("Error pruning dangling images: %v", err)
					continue
				}

				log.Infof("Pruned %d dangling images, reclaimed %d bytes", prunedImages, spaceReclaimed)

			case <-ctx.Done():
				return
			}
		}
	}()
}
