//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"os"

	"golang.org/x/sys/windows"
)

func getDiskStats(path string) (*DiskStats, error) {
	// Callers pass the POSIX root "/" (see metrics.go), which on Windows
	// resolves against the current drive of the process working directory.
	// Sample the system drive explicitly instead.
	if path == "/" || path == "\\" {
		drive := os.Getenv("SystemDrive") // e.g. "C:"
		if drive == "" {
			drive = "C:"
		}
		path = drive + `\`
	}

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	if err := windows.GetDiskFreeSpaceEx(pathPtr, &freeBytesAvailable, &totalBytes, &totalFreeBytes); err != nil {
		return nil, err
	}

	return &DiskStats{
		Total:     totalBytes,
		Used:      totalBytes - totalFreeBytes,
		Available: freeBytesAvailable,
		Path:      path,
	}, nil
}
