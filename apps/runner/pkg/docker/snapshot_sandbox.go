// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models/enums"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type captureContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var capture_context_map = cmap.New[captureContext]()

// capture_pause_map records, per sandbox, the capture context that paused the
// container for commit and therefore owes the unpause. A superseding capture
// that finds the container already paused adopts the record, so exactly one
// capture unpauses — and never one that has been superseded.
var capture_pause_map = cmap.New[context.Context]()

// claimCapturePause atomically takes the pause record for sandboxID if it is
// still held by owner. Only the claimant may unpause the container, so a
// superseded capture (whose successor adopted the record) never unpauses
// underneath the successor's commit. A nil owner marks a synchronous capture,
// which never registers a record and always owns its own pause.
func claimCapturePause(sandboxID string, owner context.Context) bool {
	if owner == nil {
		return true
	}
	return capture_pause_map.RemoveCb(sandboxID, func(_ string, v context.Context, exists bool) bool {
		return exists && v == owner
	})
}

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
	return d.createSnapshotFromSandbox(ctx, sandboxID, registry, nil)
}

// createSnapshotFromSandbox implements CreateSnapshotFromSandbox. owner is the
// context registered in capture_context_map for an async capture (nil for a
// synchronous one) and serves as the pause-ownership identity: it decides who
// is responsible for unpausing the container when captures supersede each
// other.
func (d *DockerClient) createSnapshotFromSandbox(ctx context.Context, sandboxID string, registry *dto.RegistryDTO, owner context.Context) (*dto.SnapshotInfoResponse, error) {
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
	defer func() {
		if !pausedByUs || !claimCapturePause(sandboxID, owner) {
			return
		}
		unpauseCtx, unpauseCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer unpauseCancel()
		if upErr := d.apiClient.ContainerUnpause(unpauseCtx, sandboxID); upErr != nil {
			d.logger.ErrorContext(unpauseCtx, "Failed to unpause container after snapshot", "containerId", sandboxID, "error", upErr)
		}
	}()

	switch {
	case ct.State != nil && ct.State.Running && !ct.State.Paused:
		d.logger.InfoContext(ctx, "Pausing container for consistent snapshot commit", "containerId", sandboxID)
		if err := d.apiClient.ContainerPause(ctx, sandboxID); err != nil {
			return nil, fmt.Errorf("pause container for snapshot: %w", err)
		}
		pausedByUs = true
		if owner != nil {
			capture_pause_map.Set(sandboxID, owner)
		}
	case ct.State != nil && ct.State.Paused && owner != nil:
		// The container is already paused. If a capture paused it — necessarily
		// the one this capture superseded — adopt the unpause responsibility:
		// the superseded goroutine skips its unpause once it loses the pause
		// record, so without adoption nobody would ever unpause. A container
		// paused outside any capture is left alone.
		if capture_pause_map.RemoveCb(sandboxID, func(_ string, _ context.Context, exists bool) bool { return exists }) {
			capture_pause_map.Set(sandboxID, owner)
			pausedByUs = true
			d.logger.InfoContext(ctx, "Adopted pause of superseded capture for snapshot commit", "containerId", sandboxID)
		}
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

	if pausedByUs && claimCapturePause(sandboxID, owner) {
		unpauseCtx, unpauseCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		upErr := d.apiClient.ContainerUnpause(unpauseCtx, sandboxID)
		unpauseCancel()
		if upErr != nil {
			d.logger.ErrorContext(ctx, "Unpause after snapshot commit failed", "containerId", sandboxID, "error", upErr)
			// Give the pause record back so the deferred unpause retries.
			if owner != nil {
				capture_pause_map.Set(sandboxID, owner)
			}
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

// CreateSnapshotFromSandboxAsync validates the request synchronously (the
// container must exist), records IN_PROGRESS in the snapshot-from-sandbox
// info cache, and runs the capture in a background goroutine detached from
// the request context. The capture deadline is the existing backupTimeoutMin
// applied inside CreateSnapshotFromSandbox. Progress is reported via the
// cache, queried through GET /sandboxes/{sandboxId}/snapshot-from-sandbox.
func (d *DockerClient) CreateSnapshotFromSandboxAsync(ctx context.Context, sandboxID string, request dto.CreateSnapshotFromSandboxRequestDTO) error {
	// Cancel a capture if one is already in progress for this sandbox
	prior, ok := capture_context_map.Get(sandboxID)
	if ok {
		prior.cancel()
	}

	// Inspect synchronously so a missing container surfaces as the raw docker
	// not-found error (middleware maps it to 404) before we accept the work.
	if _, err := d.ContainerInspect(ctx, sandboxID); err != nil {
		return err
	}

	d.logger.InfoContext(ctx, "Starting async snapshot capture for sandbox", "sandboxId", sandboxID, "snapshot", request.Name)

	// Register the capture context before writing IN_PROGRESS: the terminal
	// cache write of a capture this one supersedes is guarded by registration
	// ownership (under the same map shard lock), so once our registration is
	// in place the superseded goroutine can no longer clobber our entry.
	// Registering before returning also guarantees a follow-up request can
	// always cancel this capture (registering inside the goroutine would leave
	// a window where an in-flight capture is invisible to cancel-prior).
	captureCtx, cancel := context.WithCancel(context.Background())
	capture_context_map.Set(sandboxID, captureContext{captureCtx, cancel})

	// Write IN_PROGRESS before returning so the API's first poll always sees
	// an entry.
	cacheErr := d.snapshotFromSandboxInfoCache.SetCaptureState(ctx, sandboxID, request.Name, enums.SnapshotFromSandboxStateInProgress, nil, nil)
	if cacheErr != nil {
		d.logger.DebugContext(ctx, "Failed to update snapshot capture info", "error", cacheErr)
	}

	go d.captureSnapshotFromSandbox(captureCtx, cancel, sandboxID, request)

	return nil
}

func (d *DockerClient) captureSnapshotFromSandbox(ctx context.Context, cancel context.CancelFunc, sandboxID string, request dto.CreateSnapshotFromSandboxRequestDTO) {
	defer cancel()

	info, err := d.createSnapshotFromSandbox(ctx, sandboxID, request.Registry, ctx)

	state := enums.SnapshotFromSandboxStateCompleted
	switch {
	case err == nil:
		d.logger.InfoContext(ctx, "Snapshot capture completed", "sandboxId", sandboxID, "snapshot", request.Name, "imageRef", info.Name)
	case errors.Is(err, context.DeadlineExceeded):
		state = enums.SnapshotFromSandboxStateFailed
		err = fmt.Errorf("snapshot capture timed out after %dm", d.backupTimeoutMin)
		d.logger.ErrorContext(ctx, "Snapshot capture timed out", "sandboxId", sandboxID, "snapshot", request.Name)
	case errors.Is(err, context.Canceled):
		// Unlike backups (which reset to NONE on cancel), a canceled capture is
		// recorded as FAILED so the API poller observes a definite terminal
		// state instead of NONE, which it treats as "runner lost the capture".
		state = enums.SnapshotFromSandboxStateFailed
		err = errors.New("snapshot capture canceled (superseded or runner shutting down)")
		d.logger.InfoContext(ctx, "Snapshot capture canceled", "sandboxId", sandboxID, "snapshot", request.Name)
	default:
		state = enums.SnapshotFromSandboxStateFailed
		d.logger.ErrorContext(ctx, "Snapshot capture failed", "sandboxId", sandboxID, "snapshot", request.Name, "error", err)
	}

	// Record the terminal state and drop our registration in one atomic step —
	// but only while this goroutine still owns the registration. A superseding
	// capture replaces the map entry before writing its own IN_PROGRESS, so by
	// writing under the map shard lock a superseded capture can never clobber
	// the successor's cache entry with its stale terminal state.
	capture_context_map.RemoveCb(sandboxID, func(_ string, v captureContext, exists bool) bool {
		if !exists || v.ctx != ctx {
			return false
		}
		if cacheErr := d.snapshotFromSandboxInfoCache.SetCaptureState(ctx, sandboxID, request.Name, state, info, err); cacheErr != nil {
			d.logger.DebugContext(ctx, "Failed to update snapshot capture info", "error", cacheErr)
		}
		return true
	})
}
