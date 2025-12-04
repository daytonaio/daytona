// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type RunnerInfoResponseDTO struct {
	Metrics *RunnerMetrics `json:"metrics,omitempty"`
	Version string         `json:"version"`
} //	@name	RunnerInfoResponseDTO

type RunnerMetrics struct {
	CurrentCpuUsagePercentage    float64 `json:"currentCpuUsagePercentage"`
	CurrentMemoryUsagePercentage float64 `json:"currentMemoryUsagePercentage"`
	CurrentDiskUsagePercentage   float64 `json:"currentDiskUsagePercentage"`
	CurrentAllocatedCpu          float64 `json:"currentAllocatedCpu"`
	CurrentAllocatedMemoryGiB    float64 `json:"currentAllocatedMemoryGiB"`
	CurrentAllocatedDiskGiB      float64 `json:"currentAllocatedDiskGiB"`
	CurrentSnapshotCount         int64   `json:"currentSnapshotCount"`
} //	@name	RunnerMetrics
