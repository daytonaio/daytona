// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package models

import "time"

// SystemMetrics holds system-level metrics for the runner
type SystemMetrics struct {
	CPUUsage        float64   `json:"cpuUsage"`        // CPU usage percentage
	RAMUsage        float64   `json:"ramUsage"`        // RAM usage percentage
	DiskUsage       float64   `json:"diskUsage"`       // Disk usage percentage
	TotalCPU        int64     `json:"totalCPU"`        // Total CPU cores
	TotalRAMGiB     float64   `json:"totalRAMGiB"`     // Total RAM in GiB
	TotalDiskGiB    float64   `json:"totalDiskGiB"`    // Total disk in GiB
	AllocatedCPU    int64     `json:"allocatedCPU"`    // Allocated CPU cores
	AllocatedMemory int64     `json:"allocatedMemory"` // Allocated memory in GiB
	AllocatedDisk   int64     `json:"allocatedDisk"`   // Allocated disk in GB
	SnapshotCount   int64     `json:"snapshotCount"`   // Number of snapshots
	LastUpdated     time.Time `json:"lastUpdated"`     // When metrics were last updated
}
