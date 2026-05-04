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

	// Inspect services first so that failures in dependent subsystems
	// (e.g. metrics collection, which talks to Docker) do not prevent us
	// from reporting unhealthy services back to the API.
	servicesInfo := runnerInstance.InspectRunnerServices(ctx.Request.Context())

	response := dto.RunnerInfoResponseDTO{
		ServiceHealth: mapRunnerServiceInfoToDTO(servicesInfo),
		AppVersion:    internal.Version,
	}

	metrics, err := runnerInstance.MetricsCollector.Collect(ctx.Request.Context())
	if err != nil {
		// Metric collection depends on Docker; if Docker is down we still
		// want to surface the serviceHealth payload above. Log and continue
		// with a metrics-less response rather than failing the whole call.
		runnerInstance.Logger.WarnContext(ctx.Request.Context(), "Failed to collect runner metrics; returning info without metrics", "error", err)
	} else {
		response.Metrics = &dto.RunnerMetrics{
			CurrentCpuLoadAverage:        float64(metrics.CPULoadAverage),
			CurrentCpuUsagePercentage:    float64(metrics.CPUUsagePercentage),
			CurrentMemoryUsagePercentage: float64(metrics.MemoryUsagePercentage),
			CurrentDiskUsagePercentage:   float64(metrics.DiskUsagePercentage),
			CurrentAllocatedCpu:          float64(metrics.AllocatedCPU),
			CurrentAllocatedMemoryGiB:    float64(metrics.AllocatedMemoryGiB),
			CurrentAllocatedDiskGiB:      float64(metrics.AllocatedDiskGiB),
			CurrentSnapshotCount:         int(metrics.SnapshotCount),
			CurrentStartedSandboxes:      int64(metrics.StartedSandboxCount),
		}
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
