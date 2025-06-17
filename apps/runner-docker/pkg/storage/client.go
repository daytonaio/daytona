// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"
)

// ObjectStorageClient defines the interface for object storage operations
type ObjectStorageClient interface {
	GetObject(ctx context.Context, organizationId, hash string) ([]byte, error)
}
