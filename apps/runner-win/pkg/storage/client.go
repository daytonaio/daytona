// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"
	"io"
)

// ObjectStorageClient defines the interface for object storage operations
type ObjectStorageClient interface {
	// GetObject retrieves an object from storage
	GetObject(ctx context.Context, organizationId, hash string) ([]byte, error)
	// PutSnapshot uploads a snapshot file to the snapshot store (legacy, without org namespacing)
	// Returns the object path in the bucket
	// PutSnapshot uploads a snapshot file to the snapshot store (legacy, without org namespacing)
	// Returns the snapshot ref (without snapshots/ prefix), e.g., "myapp.qcow2"
	PutSnapshot(ctx context.Context, snapshotName string, reader io.Reader, size int64) (string, error)
	// PutSnapshotWithOrg uploads a snapshot file with organization namespacing
	// Returns the snapshot ref (without snapshots/ prefix), e.g., "{organizationId}/myapp.qcow2"
	PutSnapshotWithOrg(ctx context.Context, organizationId, snapshotName string, reader io.Reader, size int64) (string, error)
	// GetSnapshot retrieves a snapshot file from the snapshot store
	GetSnapshot(ctx context.Context, snapshotName string) (io.ReadCloser, int64, error)
	// DeleteSnapshot removes a snapshot from the snapshot store
	DeleteSnapshot(ctx context.Context, snapshotName string) error
	// SnapshotExists checks if a snapshot exists in the store
	SnapshotExists(ctx context.Context, snapshotName string) (bool, error)
}
