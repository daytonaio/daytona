// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"
)

// RunnerInfo 			godoc
//
//	@Summary		Runner info
//	@Description	Runner info with system metrics
//	@Produce		json
//	@Success		200	{object}	dto.RunnerInfoResponseDTO
//	@Router			/info [get]
//
//	@id				RunnerInfo
func RunnerInfo(ctx *gin.Context) {
	runnerInstance := runner.GetInstance(nil)

	// Get cached system metrics
	cpuUsage, ramUsage, diskUsage, allocatedCpu, allocatedMemory, allocatedDisk, snapshotCount := runnerInstance.MetricsService.GetCachedSystemMetrics(ctx.Request.Context())

	// Create metrics object
	metrics := &dto.RunnerMetrics{
		CurrentCpuUsagePercentage:    cpuUsage,
		CurrentMemoryUsagePercentage: ramUsage,
		CurrentDiskUsagePercentage:   diskUsage,
		CurrentAllocatedCpu:          allocatedCpu,
		CurrentAllocatedMemoryGiB:    allocatedMemory,
		CurrentAllocatedDiskGiB:      allocatedDisk,
		CurrentSnapshotCount:         snapshotCount,
	}

	response := dto.RunnerInfoResponseDTO{
		Metrics: metrics,
	}

	ctx.JSON(http.StatusOK, response)
}
