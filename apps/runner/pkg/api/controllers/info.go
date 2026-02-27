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
//	@Router			/info [get]
//
//	@id				RunnerInfo
func RunnerInfo(ctx *gin.Context) {
	runnerInstance, err := runner.GetInstance(nil)
	if err != nil {
		ctx.Error(err)
		return
	}

	metrics, err := runnerInstance.MetricsCollector.Collect(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	response := dto.RunnerInfoResponseDTO{
		Metrics: &dto.RunnerMetrics{
			CurrentCpuLoadAverage:        float64(metrics.CPULoadAverage),
			CurrentCpuUsagePercentage:    float64(metrics.CPUUsagePercentage),
			CurrentMemoryUsagePercentage: float64(metrics.MemoryUsagePercentage),
			CurrentDiskUsagePercentage:   float64(metrics.DiskUsagePercentage),
			CurrentAllocatedCpu:          int64(metrics.AllocatedCPU),
			CurrentAllocatedMemoryGiB:    int64(metrics.AllocatedMemoryGiB),
			CurrentAllocatedDiskGiB:      int64(metrics.AllocatedDiskGiB),
			CurrentSnapshotCount:         int(metrics.SnapshotCount),
			CurrentStartedSandboxes:      int64(metrics.StartedSandboxCount),
		},
		AppVersion: internal.Version,
	}

	ctx.JSON(http.StatusOK, response)
}
