// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package job

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListJobs godoc
//
//	@Tags			job
//	@Summary		List jobs
//	@Description	List jobs
//	@Produce		json
//	@Success		200	{array} Job
//	@Router			/job [get]
//
//	@id				ListJobs
func ListJobs(ctx *gin.Context) {
	server := server.GetInstance(nil)

	jobs, err := server.JobService.List(nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list jobs: %s", err.Error()))
		return
	}

	ctx.JSON(200, jobs)
}
