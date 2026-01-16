// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	"github.com/daytonaio/runner-win/pkg/storage"
	log "github.com/sirupsen/logrus"
)

// PullSnapshot downloads a snapshot from the S3-compatible snapshot store to the
// snapshots directory on the libvirt host, making it available as a base image
// for creating new sandboxes.
//
// The process:
// 1. Ensure the snapshots directory exists on the target host
// 2. Acquire a lock to prevent concurrent pulls of the same snapshot
// 3. Check if snapshot already exists on the target host (skip download if valid)
// 4. Check if snapshot exists in the remote store
// 5. Download the snapshot to the target host's snapshots directory
// 6. Set proper permissions for libvirt
// 7. Validate the downloaded image
// 8. Release the lock
//
// This function handles both local and remote libvirt hosts. When connected to a remote
// host via SSH (e.g., qemu+ssh://root@host/system), the snapshot is streamed from S3
// to the remote host via SSH.
//
// Thread-safety: This function uses file-based locking to prevent multiple concurrent
// pull operations from corrupting each other. Only one pull per snapshot can proceed
// at a time. Other callers will wait for the lock and then find the snapshot already
// exists.
func (l *LibVirt) PullSnapshot(ctx context.Context, req dto.PullSnapshotRequestDTO) error {
	snapshotName := req.Snapshot
	if snapshotName == "" {
		return fmt.Errorf("snapshot name is required")
	}

	log.Infof("PullSnapshot: Pulling snapshot '%s' from store", snapshotName)

	// Determine the path for the snapshot on the target host
	targetPath := l.getSnapshotLocalPath(snapshotName)
	tempPath := targetPath + ".downloading"

	// Ensure the snapshots directory exists on the target host BEFORE acquiring the lock
	// This is necessary because the lock file is created in the snapshots directory
	if err := l.ensureDir(ctx, snapshotsBasePath); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	// For remote libvirt hosts, also ensure the directory exists locally for lock files
	// The lock mechanism uses local files (os.OpenFile) even when libvirt is remote
	if !l.isLocalURI() {
		if err := os.MkdirAll(snapshotsBasePath, 0755); err != nil {
			return fmt.Errorf("failed to create local snapshots directory for lock files: %w", err)
		}
	}

	// Acquire lock to prevent concurrent pulls of the same snapshot
	// This is critical to prevent race conditions where multiple sandbox creations
	// try to pull the same snapshot simultaneously, causing file corruption
	releaseLock, err := l.acquireSnapshotLock(ctx, targetPath)
	if err != nil {
		return fmt.Errorf("failed to acquire snapshot lock: %w", err)
	}
	defer releaseLock()

	// Re-check if snapshot exists after acquiring lock (another process may have completed the download)
	exists, err := l.fileExists(ctx, targetPath)
	if err != nil {
		log.Warnf("PullSnapshot: Error checking if snapshot exists: %v", err)
	}
	if exists {
		log.Infof("PullSnapshot: Snapshot '%s' already exists at '%s'", snapshotName, targetPath)
		// Validate the existing image
		if err := l.runQemuImgCheck(ctx, targetPath); err != nil {
			log.Warnf("PullSnapshot: Local snapshot '%s' failed validation, re-downloading: %v", snapshotName, err)
			// Remove corrupted file and continue with download
			l.removeFile(ctx, targetPath)
		} else {
			return nil // Already exists and is valid
		}
	}

	// Also check for and clean up any stale temp files from interrupted downloads
	if tempExists, _ := l.fileExists(ctx, tempPath); tempExists {
		log.Warnf("PullSnapshot: Removing stale temp file from interrupted download: %s", tempPath)
		l.removeFile(ctx, tempPath)
	}

	// Get storage client
	storageClient, err := storage.GetObjectStorageClient()
	if err != nil {
		return fmt.Errorf("failed to get storage client: %w", err)
	}

	// Check if snapshot exists in the remote store
	storeExists, err := storageClient.SnapshotExists(ctx, snapshotName)
	if err != nil {
		return fmt.Errorf("failed to check snapshot existence in store: %w", err)
	}
	if !storeExists {
		return fmt.Errorf("snapshot '%s' not found in store", snapshotName)
	}

	// Download the snapshot from S3
	log.Infof("PullSnapshot: Downloading snapshot '%s' to '%s'", snapshotName, targetPath)

	reader, size, err := storageClient.GetSnapshot(ctx, snapshotName)
	if err != nil {
		return fmt.Errorf("failed to get snapshot from storage: %w", err)
	}
	defer reader.Close()

	// Write to the target host (local or remote)
	var written int64
	if l.isLocalURI() {
		// Local: write directly to file
		written, err = l.writeToLocalFile(ctx, reader, tempPath, snapshotName, size)
	} else {
		// Remote: stream to remote host via SSH
		log.Infof("PullSnapshot: Streaming snapshot to remote host")
		written, err = l.writeToRemoteFile(ctx, reader, tempPath, snapshotName, size)
	}

	if err != nil {
		l.removeFile(ctx, tempPath)
		return fmt.Errorf("failed to download snapshot: %w", err)
	}

	if written != size {
		l.removeFile(ctx, tempPath)
		return fmt.Errorf("incomplete download: expected %d bytes, got %d", size, written)
	}

	// Move temp file to final location
	if err := l.renameFile(ctx, tempPath, targetPath); err != nil {
		l.removeFile(ctx, tempPath)
		return fmt.Errorf("failed to finalize snapshot file: %w", err)
	}

	// Set proper permissions for libvirt
	if err := l.chmodFile(ctx, targetPath, 0644); err != nil {
		log.Warnf("PullSnapshot: Failed to set permissions on %s: %v", targetPath, err)
	}

	// Change ownership to libvirt-qemu:kvm (best effort)
	if err := l.chownLibvirt(ctx, targetPath); err != nil {
		log.Warnf("PullSnapshot: Failed to set ownership on %s: %v", targetPath, err)
	}

	// Validate the downloaded image
	if err := l.runQemuImgCheck(ctx, targetPath); err != nil {
		l.removeFile(ctx, targetPath)
		return fmt.Errorf("downloaded snapshot failed validation: %w", err)
	}

	log.Infof("PullSnapshot: Successfully pulled snapshot '%s' (%d bytes) to '%s'", snapshotName, written, targetPath)
	return nil
}

// writeToLocalFile writes data from a reader to a local file with progress logging
func (l *LibVirt) writeToLocalFile(ctx context.Context, reader io.Reader, path, name string, totalSize int64) (int64, error) {
	file, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Wrap with progress logging
	progressReader := &progressReaderWithLog{
		reader:       reader,
		total:        totalSize,
		snapshotName: name,
		logInterval:  100 * 1024 * 1024, // 100MB
	}

	return io.Copy(file, progressReader)
}

// writeToRemoteFile streams data from a reader to a remote file via SSH
func (l *LibVirt) writeToRemoteFile(ctx context.Context, reader io.Reader, path, name string, totalSize int64) (int64, error) {
	writer, err := l.openRemoteFileForWrite(ctx, path)
	if err != nil {
		return 0, fmt.Errorf("failed to open remote file: %w", err)
	}

	// Wrap reader with progress logging
	progressReader := &progressReaderWithLog{
		reader:       reader,
		total:        totalSize,
		snapshotName: name,
		logInterval:  100 * 1024 * 1024, // 100MB
	}

	written, copyErr := io.Copy(writer, progressReader)
	closeErr := writer.Close()

	if copyErr != nil {
		return written, copyErr
	}
	if closeErr != nil {
		return written, closeErr
	}

	return written, nil
}

// progressReaderWithLog wraps an io.Reader to log download progress
type progressReaderWithLog struct {
	reader       io.Reader
	total        int64
	read         int64
	snapshotName string
	lastLog      int64
	logInterval  int64
}

func (pr *progressReaderWithLog) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.read += int64(n)

	if pr.read-pr.lastLog >= pr.logInterval || (err == io.EOF && pr.read > 0) {
		percent := float64(pr.read) / float64(pr.total) * 100
		log.Infof("PullSnapshot: Downloading '%s': %.1f%% (%d / %d bytes)",
			pr.snapshotName, percent, pr.read, pr.total)
		pr.lastLog = pr.read
	}

	return n, err
}

// getSnapshotLocalPath returns the path for a snapshot on the target host
func (l *LibVirt) getSnapshotLocalPath(snapshotName string) string {
	// Strip "snapshots/" prefix if present (snapshot.ref contains S3 path like "snapshots/name.qcow2")
	snapshotName = strings.TrimPrefix(snapshotName, "snapshots/")

	// Ensure .qcow2 extension
	if !strings.HasSuffix(snapshotName, ".qcow2") {
		snapshotName = snapshotName + ".qcow2"
	}
	return filepath.Join(snapshotsBasePath, snapshotName)
}

// ListLocalSnapshots returns a list of snapshots available in the snapshots directory
// on the target host (local or remote)
func (l *LibVirt) ListLocalSnapshots(ctx context.Context) ([]string, error) {
	names, err := l.readDir(ctx, snapshotsBasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	var snapshots []string
	for _, name := range names {
		if strings.HasSuffix(name, ".qcow2") {
			// Return name without extension
			snapshots = append(snapshots, strings.TrimSuffix(name, ".qcow2"))
		}
	}

	return snapshots, nil
}

// GetLocalSnapshotInfo returns information about a snapshot on the target host
func (l *LibVirt) GetLocalSnapshotInfo(ctx context.Context, snapshotName string) (*LocalSnapshotInfo, error) {
	targetPath := l.getSnapshotLocalPath(snapshotName)

	exists, err := l.fileExists(ctx, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check snapshot: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("snapshot '%s' not found", snapshotName)
	}

	size, err := l.getFileSize(ctx, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot size: %w", err)
	}

	return &LocalSnapshotInfo{
		Name:        snapshotName,
		Path:        targetPath,
		SizeBytes:   size,
		VirtualSize: 0, // Would require parsing qemu-img info output
		ModTime:     nil,
	}, nil
}

// LocalSnapshotInfo contains information about a snapshot on the target host
type LocalSnapshotInfo struct {
	Name        string
	Path        string
	SizeBytes   int64
	VirtualSize int64
	ModTime     interface{} // time.Time for local, nil for remote
}

// DeleteLocalSnapshot removes a snapshot from the snapshots directory on the target host
func (l *LibVirt) DeleteLocalSnapshot(ctx context.Context, snapshotName string) error {
	targetPath := l.getSnapshotLocalPath(snapshotName)

	exists, err := l.fileExists(ctx, targetPath)
	if err != nil {
		return fmt.Errorf("failed to check snapshot: %w", err)
	}
	if !exists {
		return fmt.Errorf("snapshot '%s' not found", snapshotName)
	}

	if err := l.removeFile(ctx, targetPath); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	log.Infof("DeleteLocalSnapshot: Deleted snapshot '%s' from '%s'", snapshotName, targetPath)
	return nil
}
