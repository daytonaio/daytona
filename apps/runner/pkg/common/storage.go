// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"

	"github.com/docker/go-units"
)

// ParseStorageOptSizeGB parses the size value from Docker StorageOpt["size"] and returns GB as float64.
// Docker storage-opt uses binary units (GiB) for both string ("3G") and numeric formats.
// Supports formats like "10G", "10240M", "10737418240" (bytes), etc.
func ParseStorageOptSizeGB(storageOpt map[string]string) (float64, error) {
	if storageOpt == nil {
		return 0, fmt.Errorf("storageOpt is nil")
	}

	sizeStr, ok := storageOpt["size"]
	if !ok {
		return 0, fmt.Errorf("size not found in storageOpt")
	}

	// Parse size string using docker units library which handles various formats
	sizeBytes, err := units.RAMInBytes(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse storage size '%s': %w", sizeStr, err)
	}

	// Convert bytes to GB
	return float64(sizeBytes) / (1024 * 1024 * 1024), nil
}

// GBToBytes converts gigabytes to bytes
func GBToBytes(gb float64) int64 {
	return int64(gb * 1024 * 1024 * 1024)
}
