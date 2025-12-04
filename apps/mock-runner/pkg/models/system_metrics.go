// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package models

type SystemMetrics struct {
	CPUUsage        float64
	RAMUsage        float64
	DiskUsage       float64
	AllocatedCPU    float64
	AllocatedMemory float64
	AllocatedDisk   float64
	SnapshotCount   int64
}
