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
	metrics := runnerInstance.MetricsService.GetSystemMetrics(ctx.Request.Context())

	response := dto.RunnerInfoResponseDTO{
		Metrics: &dto.RunnerMetrics{
			CurrentCpuUsagePercentage:    metrics.CPUUsage,
			CurrentMemoryUsagePercentage: metrics.RAMUsage,
			CurrentDiskUsagePercentage:   metrics.DiskUsage,
			CurrentAllocatedCpu:          metrics.AllocatedCPU,
			CurrentAllocatedMemoryGiB:    metrics.AllocatedMemory,
			CurrentAllocatedDiskGiB:      metrics.AllocatedDisk,
			CurrentSnapshotCount:         metrics.SnapshotCount,
		},
	}

	ctx.JSON(http.StatusOK, response)
}
