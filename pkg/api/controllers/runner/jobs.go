// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

const LONG_POLL_TIMEOUT = 1 * time.Minute
const LONG_POLL_INTERVAL = 50 * time.Millisecond

// ListRunnerJobs 			godoc
//
//	@Tags			runner
//	@Summary		List runner jobs
//	@Description	List runner jobs
//	@Param			runnerId	path	string	true	"Runner ID"
//	@Produce		json
//	@Success		200	{array}	Job
//	@Router			/runner/{runnerId}/jobs [get]
//
//	@id				ListRunnerJobs
func ListRunnerJobs(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")

	server := server.GetInstance(nil)

	timeout := time.After(LONG_POLL_TIMEOUT)
	ticker := time.NewTicker(LONG_POLL_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Request.Context().Done():
			// Handle client cancelling the request
			ctx.AbortWithStatus(http.StatusRequestTimeout)
			return
		case <-timeout:
			// Handle request timing out
			ctx.JSON(http.StatusNoContent, nil)
			return
		case <-ticker.C:
			// Check for new jobs
			jobs, err := server.RunnerService.ListRunnerJobs(ctx.Request.Context(), runnerId)
			if err != nil {
				ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get runner jobs: %w", err))
				return
			}
			if len(jobs) > 0 {
				ctx.JSON(http.StatusOK, jobs)
				return
			}
		}
	}
}

// UpdateJobState 			godoc
//
//	@Tags			runner
//	@Summary		Update job state
//	@Description	Update job state
//	@Param			runnerId		path	string			true	"Runner ID"
//	@Param			jobId			path	string			true	"Job ID"
//	@Param			updateJobState	body	UpdateJobState	true	"Update job state"
//	@Produce		json
//	@Success		200
//	@Router			/runner/{runnerId}/jobs/{jobId}/state [post]
//
//	@id				UpdateJobState
func UpdateJobState(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")
	jobId := ctx.Param("jobId")

	var updateJobState services.UpdateJobStateDTO
	err := ctx.BindJSON(&updateJobState)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	job, err := server.JobService.Find(ctx.Request.Context(), &stores.JobFilter{
		Id: &jobId,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get job: %w", err))
		return
	}

	if job.RunnerId != nil && *job.RunnerId != runnerId {
		ctx.AbortWithError(http.StatusUnauthorized, fmt.Errorf("job does not belong to runner"))
		return
	}

	err = server.RunnerService.UpdateJobState(ctx.Request.Context(), jobId, updateJobState)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to update job state: %w", err))
		return
	}

	ctx.Status(200)
}
