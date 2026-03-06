// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"

	"github.com/daytonaio/runner/internal"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models"
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

	servicesInfo := runnerInstance.InspectRunnerServices(ctx.Request.Context())

	response := dto.RunnerInfoResponseDTO{
		ServiceHealth: mapRunnerServiceInfoToDTO(servicesInfo),
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

func mapRunnerServiceInfoToDTO(servicesInfo []models.RunnerServiceInfo) []*dto.RunnerServiceInfo {
	runnerServicesInfoDTO := make([]*dto.RunnerServiceInfo, 0)

	for _, serviceInfo := range servicesInfo {
		serviceInfoDto := &dto.RunnerServiceInfo{
			ServiceName: serviceInfo.ServiceName,
			Healthy:     serviceInfo.Healthy,
		}

		if serviceInfo.Err != nil {
			errReason := serviceInfo.Err.Error()
			serviceInfoDto.ErrorReason = &errReason
		}

		runnerServicesInfoDTO = append(runnerServicesInfoDTO, serviceInfoDto)
	}

	return runnerServicesInfoDTO
}
