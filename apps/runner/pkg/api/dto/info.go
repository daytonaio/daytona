// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type RunnerMetrics struct {
	CurrentCpuLoadAverage        float64 `json:"currentCpuLoadAverage"`
	CurrentCpuUsagePercentage    float64 `json:"currentCpuUsagePercentage"`
	CurrentMemoryUsagePercentage float64 `json:"currentMemoryUsagePercentage"`
	CurrentDiskUsagePercentage   float64 `json:"currentDiskUsagePercentage"`
	CurrentAllocatedCpu          int64   `json:"currentAllocatedCpu"`
	CurrentAllocatedMemoryGiB    int64   `json:"currentAllocatedMemoryGiB"`
	CurrentAllocatedDiskGiB      int64   `json:"currentAllocatedDiskGiB"`
	CurrentSnapshotCount         int     `json:"currentSnapshotCount"`
	CurrentStartedSandboxes      int64   `json:"currentStartedSandboxes"`
} //	@name	RunnerMetrics

type RunnerInfoResponseDTO struct {
	Metrics    *RunnerMetrics `json:"metrics,omitempty"`
	AppVersion string         `json:"appVersion"`
} //	@name	RunnerInfoResponseDTO
