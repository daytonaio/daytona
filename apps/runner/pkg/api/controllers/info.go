// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"

	"github.com/daytonaio/runner/internal"
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
//	@Failure		500	{object}	gin.Error
//	@Router			/info [get]
//
//	@id				RunnerInfo
func RunnerInfo(ctx *gin.Context) {
	runnerInstance := runner.GetInstance(nil)

	// Get cached system metrics
	metrics, err := runnerInstance.MetricsService.GetMetrics()
	if err != nil {
		ctx.Error(err)
		return
	}

	response := dto.RunnerInfoResponseDTO{
		Metrics: &dto.RunnerMetrics{
			CurrentCpuUsagePercentage:    &metrics.CPUUsage,
			CurrentCpuLoadAverage:        &metrics.CPULoadAvg,
			CurrentMemoryUsagePercentage: &metrics.RAMUsage,
			CurrentDiskUsagePercentage:   &metrics.DiskUsage,
			CurrentAllocatedCpu:          &metrics.AllocatedCPU,
			CurrentAllocatedMemoryGiB:    &metrics.AllocatedMemory,
			CurrentAllocatedDiskGiB:      &metrics.AllocatedDisk,
			CurrentSnapshotCount:         &metrics.SnapshotCount,
		},
		Version: internal.Version,
	}

	ctx.JSON(http.StatusOK, response)
}
