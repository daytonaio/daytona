// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package metrics

// ResourceMetrics represents the current resource usage metrics as percentages
type ResourceMetrics struct {
	CPUPercentage    float64 `json:"cpuPercentage"`
	MemoryPercentage float64 `json:"memoryPercentage"`
	DiskPercentage   float64 `json:"diskPercentage"`
}

// MetricsCollector defines the interface for collecting sandbox metrics
type MetricsCollector interface {
	GetCPUPercentage() (float64, error)
	GetMemoryPercentage() (float64, error)
	GetDiskPercentage() (float64, error)
}
