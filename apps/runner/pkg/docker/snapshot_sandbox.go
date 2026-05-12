// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/daytonaio/runner/pkg/api/dto"
)

func snapshotRegistryProject(reg *dto.RegistryDTO) string {
	if reg.Project != nil && *reg.Project != "" {
		return *reg.Project
	}
	return "daytona"
}

func snapshotRegistryHost(reg *dto.RegistryDTO) string {
	host := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(reg.Url), "https://"), "http://")
	return strings.TrimSuffix(host, "/")
}

// snapshotTempImageRef is a local-only intermediate tag we commit to before
// we know the image hash. It is never pushed to the registry.
func snapshotTempImageRef(sandboxID string) string {
	return fmt.Sprintf("daytona-from-sandbox-%s:%d", sandboxID, time.Now().UnixNano())
}

func snapshotCanonicalImageRef(reg *dto.RegistryDTO, hash string) string {
	return fmt.Sprintf(
		"%s/%s/daytona-%s:daytona",
		snapshotRegistryHost(reg),
		snapshotRegistryProject(reg),
		dto.HashWithoutPrefix(hash),
	)
}

// CreateSnapshotFromSandbox commits the sandbox container filesystem, pushes
// the resulting image to the internal registry under the canonical
// `daytona-{hash}:daytona` tag, and returns image metadata for the API
// snapshot record. Running containers are briefly paused during commit to
// produce a consistent on-disk snapshot.
func (d *DockerClient) CreateSnapshotFromSandbox(ctx context.Context, sandboxID string, registry *dto.RegistryDTO) (*dto.SnapshotInfoResponse, error) {
	if registry == nil || strings.TrimSpace(registry.Url) == "" {
		return nil, fmt.Errorf("registry is required for sandbox snapshot")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(d.backupTimeoutMin)*time.Minute)
	defer cancel()

	ct, err := d.ContainerInspect(ctx, sandboxID)
	if err != nil {
		return nil, err
	}

	pausedByUs := false
	if ct.State != nil && ct.State.Running && !ct.State.Paused {
		d.logger.InfoContext(ctx, "Pausing container for consistent snapshot commit", "containerId", sandboxID)
		if err := d.apiClient.ContainerPause(ctx, sandboxID); err != nil {
			return nil, fmt.Errorf("pause container for snapshot: %w", err)
		}
		pausedByUs = true
		defer func() {
			if !pausedByUs {
				return
			}
			unpauseCtx, unpauseCancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer unpauseCancel()
			if upErr := d.apiClient.ContainerUnpause(unpauseCtx, sandboxID); upErr != nil {
				d.logger.ErrorContext(unpauseCtx, "Failed to unpause container after snapshot", "containerId", sandboxID, "error", upErr)
			}
		}()
	}

	tempRef := snapshotTempImageRef(sandboxID)

	if err := d.commitContainer(ctx, sandboxID, tempRef); err != nil {
		return nil, err
	}

	// Best-effort cleanup of the temp tag at the end. The canonical tag is
	// removed separately after push.
	defer func() {
		if rmErr := d.RemoveImage(context.Background(), tempRef, true); rmErr != nil {
			d.logger.WarnContext(ctx, "Failed to remove local temp snapshot image", "imageRef", tempRef, "error", rmErr)
		}
	}()

	if pausedByUs {
		unpauseCtx, unpauseCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		upErr := d.apiClient.ContainerUnpause(unpauseCtx, sandboxID)
		unpauseCancel()
		if upErr != nil {
			d.logger.ErrorContext(ctx, "Unpause after snapshot commit failed", "containerId", sandboxID, "error", upErr)
		} else {
			pausedByUs = false
		}
	}

	// Inspect the freshly committed image to get its content hash. For a fresh
	// commit there are no RepoDigests yet, so GetImageInfo falls back to the
	// image ID (sha256:...) which is what we want for the canonical tag.
	tempInfo, err := d.GetImageInfo(ctx, tempRef)
	if err != nil {
		return nil, fmt.Errorf("inspect committed snapshot image: %w", err)
	}

	hashNoPrefix := dto.HashWithoutPrefix(tempInfo.Hash)
	if hashNoPrefix == "" {
		return nil, fmt.Errorf("committed snapshot image has no hash")
	}

	canonicalRef := snapshotCanonicalImageRef(registry, tempInfo.Hash)

	if err := d.TagImage(ctx, tempRef, canonicalRef); err != nil {
		return nil, fmt.Errorf("tag committed snapshot image as %s: %w", canonicalRef, err)
	}

	pushedOK := false
	defer func() {
		// Always remove the canonical local tag - it's only needed long enough
		// to push. Force-remove because there may be no other tag pointing at
		// the image after the temp tag is gone.
		if rmErr := d.RemoveImage(context.Background(), canonicalRef, true); rmErr != nil && pushedOK {
			d.logger.WarnContext(ctx, "Failed to remove local canonical snapshot image after push", "imageRef", canonicalRef, "error", rmErr)
		}
	}()

	if err := d.PushImage(ctx, canonicalRef, registry); err != nil {
		return nil, err
	}
	pushedOK = true

	return &dto.SnapshotInfoResponse{
		Name:       canonicalRef,
		SizeGB:     float64(tempInfo.Size) / (1024 * 1024 * 1024),
		Entrypoint: tempInfo.Entrypoint,
		Cmd:        tempInfo.Cmd,
		Hash:       hashNoPrefix,
	}, nil
}
