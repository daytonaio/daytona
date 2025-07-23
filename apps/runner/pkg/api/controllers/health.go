// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"
)

// HealthCheck 			godoc
//
//	@Summary		Health check
//	@Description	Health check with system metrics
//	@Produce		json
//	@Success		200	{object}	dto.HealthCheckResponseDTO
//	@Router			/ [get]
//
//	@id				HealthCheck
func HealthCheck(ctx *gin.Context) {
	runnerInstance := runner.GetInstance(nil)

	// Get cached system metrics
	cpuUsage, ramUsage, diskUsage, allocatedCpu, allocatedMemory, allocatedDisk, snapshotCount := runnerInstance.MetricsService.GetCachedSystemMetrics(ctx.Request.Context())

	// Create metrics object
	metrics := &dto.HealthMetrics{
		CurrentCpuUsagePercentage:    cpuUsage,
		CurrentMemoryUsagePercentage: ramUsage,
		CurrentDiskUsagePercentage:   diskUsage,
		CurrentAllocatedCpu:          allocatedCpu,
		CurrentAllocatedMemoryGiB:    allocatedMemory,
		CurrentAllocatedDiskGiB:      allocatedDisk,
		CurrentSnapshotCount:         snapshotCount,
	}

	response := dto.HealthCheckResponseDTO{
		Status:  "ok",
		Version: "0.0.1",
		Metrics: metrics,
	}

	ctx.JSON(http.StatusOK, response)
}
