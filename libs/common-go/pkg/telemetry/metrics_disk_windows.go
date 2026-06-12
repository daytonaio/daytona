//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"os"
	"syscall"
	"unsafe"
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

	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return nil, err
	}
	proc, err := kernel32.FindProc("GetDiskFreeSpaceExW")
	if err != nil {
		return nil, err
	}

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	r1, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if r1 == 0 {
		return nil, callErr
	}

	return &DiskStats{
		Total:     totalBytes,
		Used:      totalBytes - totalFreeBytes,
		Available: freeBytesAvailable,
		Path:      path,
	}, nil
}
