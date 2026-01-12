// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	"github.com/daytonaio/runner-win/pkg/storage"
	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

// CreateSnapshot creates a snapshot from a sandbox's disk and uploads it to the snapshot store.
// Unlike PushSnapshot, this function can work with running VMs:
//   - live=false (default): Pauses the VM, flattens disk, resumes VM, then uploads to S3
//   - live=true: Uses qemu-img --force-share to read disk while VM runs (optimistic, no pause)
//
// The VM is resumed as soon as possible after disk flattening to minimize downtime.
// The S3 upload happens after the VM is back online.
//
// This function handles both local and remote libvirt hosts.
func (l *LibVirt) CreateSnapshot(ctx context.Context, req dto.CreateSnapshotRequestDTO) (*dto.CreateSnapshotResponseDTO, error) {
	log.Infof("CreateSnapshot: Creating snapshot '%s' from sandbox '%s' (live=%v)", req.Name, req.SandboxId, req.Live)

	conn, err := l.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Look up the domain
	domain, err := l.LookupDomainBySandboxId(conn, req.SandboxId)
	if err != nil {
		return nil, fmt.Errorf("sandbox not found: %w", err)
	}
	defer domain.Free()

	// Get original domain state
	originalState, _, err := domain.GetState()
	if err != nil {
		return nil, fmt.Errorf("failed to get domain state: %w", err)
	}

	wasRunning := (originalState == libvirt.DOMAIN_RUNNING)
	didPause := false

	// If not live mode and VM is running, pause it for consistent snapshot
	if !req.Live && wasRunning {
		log.Infof("CreateSnapshot: Pausing VM '%s' for consistent snapshot", req.SandboxId)
		if err := domain.Suspend(); err != nil {
			return nil, fmt.Errorf("failed to pause VM: %w", err)
		}
		didPause = true

		// Wait for the VM to fully pause and release disk locks
		// Verify the domain is actually paused
		for i := 0; i < 10; i++ {
			state, _, err := domain.GetState()
			if err != nil {
				log.Warnf("CreateSnapshot: Failed to get domain state after suspend: %v", err)
				break
			}
			if state == libvirt.DOMAIN_PAUSED {
				log.Infof("CreateSnapshot: VM '%s' is now paused", req.SandboxId)
				break
			}
			log.Debugf("CreateSnapshot: Waiting for VM to pause (state=%d, attempt=%d)", state, i+1)
			time.Sleep(500 * time.Millisecond)
		}

		// Additional delay to ensure disk locks are released
		time.Sleep(2 * time.Second)
	}

	// Get the sandbox disk path
	sandboxDiskPath := filepath.Join(sandboxesBasePath, fmt.Sprintf("%s.qcow2", req.SandboxId))

	// Temp path for the flattened snapshot
	tempSnapshotPath := fmt.Sprintf("/tmp/%s-snapshot.qcow2", req.Name)

	// Flatten the disk (with or without --force-share)
	log.Infof("CreateSnapshot: Flattening sandbox disk to %s (live=%v)", tempSnapshotPath, req.Live)
	var flattenErr error
	if req.Live {
		flattenErr = l.flattenDiskLive(ctx, sandboxDiskPath, tempSnapshotPath)
	} else {
		flattenErr = l.flattenDisk(ctx, sandboxDiskPath, tempSnapshotPath)
	}

	// Resume VM immediately after flattening (before upload) - only if we paused it
	if didPause {
		log.Infof("CreateSnapshot: Resuming VM '%s' after disk flatten", req.SandboxId)
		if resumeErr := domain.Resume(); resumeErr != nil {
			log.Errorf("CreateSnapshot: Failed to resume VM after flatten: %v", resumeErr)
			// Continue with error handling - we still want to clean up
		}
	}

	// Now handle any flatten error
	if flattenErr != nil {
		l.removeFile(ctx, tempSnapshotPath) // Best effort cleanup
		return nil, fmt.Errorf("failed to flatten disk: %w", flattenErr)
	}

	// Ensure cleanup of temp file on the target host
	defer func() {
		if err := l.removeFile(ctx, tempSnapshotPath); err != nil {
			log.Warnf("CreateSnapshot: Failed to clean up temp file %s: %v", tempSnapshotPath, err)
		}
	}()

	// Get the flattened image size
	snapshotSize, err := l.getFileSize(ctx, tempSnapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot file info: %w", err)
	}

	log.Infof("CreateSnapshot: Uploading snapshot '%s' (%d bytes) to object storage", req.Name, snapshotSize)

	// Get storage client
	storageClient, err := storage.GetObjectStorageClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage client: %w", err)
	}

	// Upload the snapshot to S3
	var objectPath string
	if l.isLocalURI() {
		// Local: open file directly
		snapshotFile, err := os.Open(tempSnapshotPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open snapshot file: %w", err)
		}
		defer snapshotFile.Close()

		objectPath, err = storageClient.PutSnapshot(ctx, req.Name, snapshotFile, snapshotSize)
		if err != nil {
			return nil, fmt.Errorf("failed to upload snapshot: %w", err)
		}
	} else {
		// Remote: stream file from remote host via SSH
		log.Infof("CreateSnapshot: Streaming snapshot from remote host")
		reader, err := l.openRemoteFileForRead(ctx, tempSnapshotPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open remote snapshot file: %w", err)
		}
		defer reader.Close()

		objectPath, err = storageClient.PutSnapshot(ctx, req.Name, reader, snapshotSize)
		if err != nil {
			return nil, fmt.Errorf("failed to upload snapshot: %w", err)
		}
	}

	log.Infof("CreateSnapshot: Successfully uploaded snapshot '%s' to '%s'", req.Name, objectPath)

	return &dto.CreateSnapshotResponseDTO{
		Name:         req.Name,
		SnapshotPath: objectPath,
		SizeBytes:    snapshotSize,
		LiveMode:     req.Live,
	}, nil
}

// flattenDiskLive converts an overlay qcow2 disk into a standalone image using --force-share.
// This allows reading the disk while the VM is running (optimistic snapshot).
// WARNING: This may produce an inconsistent snapshot if the VM is actively writing to disk.
func (l *LibVirt) flattenDiskLive(ctx context.Context, sourcePath, destPath string) error {
	isLocal := l.isLocalURI()

	// qemu-img convert -U -O qcow2 source.qcow2 dest.qcow2
	// -U (--force-share) allows reading the disk even if it's in use
	convertCmd := fmt.Sprintf("qemu-img convert -U -O qcow2 %s %s", sourcePath, destPath)

	var cmd *exec.Cmd
	if isLocal {
		log.Debugf("Executing locally (live mode): %s", convertCmd)
		cmd = exec.CommandContext(ctx, "bash", "-c", convertCmd)
	} else {
		host := l.extractHostFromURI()
		if host == "" {
			return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
		}
		log.Debugf("Executing on remote server %s (live mode): %s", host, convertCmd)
		cmd = exec.CommandContext(ctx, "ssh", host, convertCmd)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img convert (live) failed: %w (output: %s)", err, string(output))
	}

	return nil
}
