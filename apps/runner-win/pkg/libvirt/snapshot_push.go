// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	"github.com/daytonaio/runner-win/pkg/storage"
	log "github.com/sirupsen/logrus"
)

// PushSnapshot creates a new snapshot from a sandbox's disk and uploads it to the snapshot store.
// The process:
// 1. Ensure the sandbox is stopped (or pause it temporarily)
// 2. Create a standalone qcow2 image by flattening the overlay (qemu-img convert)
// 3. Upload the flattened image to S3-compatible storage
// 4. Return snapshot metadata
//
// This function handles both local and remote libvirt hosts. When connected to a remote
// host via SSH (e.g., qemu+ssh://root@host/system), all file operations are executed
// on the remote host, and the flattened image is streamed to the runner for S3 upload.
func (l *LibVirt) PushSnapshot(ctx context.Context, req dto.PushSnapshotRequestDTO) (*dto.PushSnapshotResponseDTO, error) {
	log.Infof("PushSnapshot: Creating snapshot '%s' from sandbox '%s'", req.SnapshotName, req.SandboxId)

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

	// Get domain state - we need it to be stopped or paused for a consistent snapshot
	state, _, err := domain.GetState()
	if err != nil {
		return nil, fmt.Errorf("failed to get domain state: %w", err)
	}

	// Check if domain is in a safe state for snapshotting
	if state != 5 && state != 3 { // DOMAIN_SHUTOFF=5, DOMAIN_PAUSED=3
		return nil, fmt.Errorf("sandbox must be stopped or paused to create a snapshot (current state: %d)", state)
	}

	// Get the sandbox disk path (on local or remote host)
	sandboxDiskPath := filepath.Join(sandboxesBasePath, fmt.Sprintf("%s.qcow2", req.SandboxId))

	// Temp path for the flattened snapshot (on the same host as the sandbox disk)
	tempSnapshotPath := fmt.Sprintf("/tmp/%s-snapshot.qcow2", req.SnapshotName)

	// Flatten the overlay disk into a standalone qcow2 image
	// This runs on the local or remote host depending on the libvirt URI
	log.Infof("PushSnapshot: Flattening sandbox disk to %s", tempSnapshotPath)
	if err := l.flattenDisk(ctx, sandboxDiskPath, tempSnapshotPath); err != nil {
		return nil, fmt.Errorf("failed to flatten disk: %w", err)
	}

	// Ensure cleanup of temp file on the target host
	defer func() {
		if err := l.removeFile(ctx, tempSnapshotPath); err != nil {
			log.Warnf("PushSnapshot: Failed to clean up temp file %s: %v", tempSnapshotPath, err)
		}
	}()

	// Get the flattened image size (on local or remote host)
	snapshotSize, err := l.getFileSize(ctx, tempSnapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot file info: %w", err)
	}

	log.Infof("PushSnapshot: Uploading snapshot '%s' (%d bytes) to object storage", req.SnapshotName, snapshotSize)

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

		objectPath, err = storageClient.PutSnapshot(ctx, req.SnapshotName, snapshotFile, snapshotSize)
		if err != nil {
			return nil, fmt.Errorf("failed to upload snapshot: %w", err)
		}
	} else {
		// Remote: stream file from remote host via SSH
		log.Infof("PushSnapshot: Streaming snapshot from remote host")
		reader, err := l.openRemoteFileForRead(ctx, tempSnapshotPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open remote snapshot file: %w", err)
		}
		defer reader.Close()

		objectPath, err = storageClient.PutSnapshot(ctx, req.SnapshotName, reader, snapshotSize)
		if err != nil {
			return nil, fmt.Errorf("failed to upload snapshot: %w", err)
		}
	}

	log.Infof("PushSnapshot: Successfully uploaded snapshot '%s' to '%s'", req.SnapshotName, objectPath)

	return &dto.PushSnapshotResponseDTO{
		SnapshotName: req.SnapshotName,
		SnapshotPath: objectPath,
		SizeBytes:    snapshotSize,
	}, nil
}

// flattenDisk converts an overlay qcow2 disk (with backing file) into a standalone qcow2 image.
// This is done using qemu-img convert which reads the entire chain and writes a new independent image.
// The command runs on the local or remote host depending on the libvirt URI.
func (l *LibVirt) flattenDisk(ctx context.Context, sourcePath, destPath string) error {
	isLocal := l.isLocalURI()

	// qemu-img convert -O qcow2 source.qcow2 dest.qcow2
	// This reads the entire backing chain and creates a standalone image
	convertCmd := fmt.Sprintf("qemu-img convert -O qcow2 %s %s", sourcePath, destPath)

	var cmd *exec.Cmd
	if isLocal {
		log.Debugf("Executing locally: %s", convertCmd)
		cmd = exec.CommandContext(ctx, "bash", "-c", convertCmd)
	} else {
		host := l.extractHostFromURI()
		if host == "" {
			return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
		}
		log.Debugf("Executing on remote server %s: %s", host, convertCmd)
		cmd = exec.CommandContext(ctx, "ssh", host, convertCmd)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img convert failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// PushSnapshotFromFile uploads an existing qcow2 file to the snapshot store.
// This is useful for uploading pre-built base images.
// Note: This function expects the file to be on the local runner machine.
func (l *LibVirt) PushSnapshotFromFile(ctx context.Context, snapshotName string, filePath string) (*dto.PushSnapshotResponseDTO, error) {
	log.Infof("PushSnapshotFromFile: Uploading '%s' as snapshot '%s'", filePath, snapshotName)

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Get storage client
	storageClient, err := storage.GetObjectStorageClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage client: %w", err)
	}

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Upload to object storage
	objectPath, err := storageClient.PutSnapshot(ctx, snapshotName, file, fileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to upload snapshot: %w", err)
	}

	log.Infof("PushSnapshotFromFile: Successfully uploaded '%s' to '%s'", snapshotName, objectPath)

	return &dto.PushSnapshotResponseDTO{
		SnapshotName: snapshotName,
		SnapshotPath: objectPath,
		SizeBytes:    fileSize,
	}, nil
}

// CopySnapshotToLocal downloads a snapshot from the store to the local snapshots directory.
// This is the inverse of PushSnapshot - it retrieves a snapshot for use as a base image.
// Deprecated: Use PullSnapshot with dto.PullSnapshotRequestDTO instead for more robust handling.
func (l *LibVirt) CopySnapshotToLocal(ctx context.Context, snapshotName string) (string, error) {
	// Delegate to PullSnapshot for consistent handling
	req := dto.PullSnapshotRequestDTO{
		Snapshot: snapshotName,
	}

	if err := l.PullSnapshot(ctx, req); err != nil {
		return "", err
	}

	return l.getSnapshotLocalPath(snapshotName), nil
}
