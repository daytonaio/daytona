// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package models

import (
	"time"

	"github.com/daytonaio/runner/pkg/models/enums"
)

type SystemMetrics struct {
	CPUUsage        float64   `json:"cpu_usage"`
	RAMUsage        float64   `json:"ram_usage"`
	DiskUsage       float64   `json:"disk_usage"`
	AllocatedCPU    int64     `json:"allocated_cpu"`
	AllocatedMemory int64     `json:"allocated_memory"`
	AllocatedDisk   int64     `json:"allocated_disk"`
	SnapshotCount   int       `json:"snapshot_count"`
	LastUpdated     time.Time `json:"last_updated"`
}

type CacheData struct {
	SandboxState    enums.SandboxState
	BackupState     enums.BackupState
	DestructionTime *time.Time
	SystemMetrics   *SystemMetrics
}
