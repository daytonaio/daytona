// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package job

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// ListJobs godoc
//
//	@Tags			job
//	@Summary		List jobs
//	@Description	List jobs
//	@Param			states	query	[]string	false	"Job states"
//	@Produce		json
//	@Success		200	{array}	Job
//	@Router			/job [get]
//
//	@id				ListJobs
func ListJobs(ctx *gin.Context) {
	states := ctx.QueryArray("states")

	server := server.GetInstance(nil)

	jobStates := util.ArrayMap(states, func(s string) models.JobState {
		return models.JobState(s)
	})

	jobs, err := server.JobService.List(ctx.Request.Context(), &stores.JobFilter{
		States: &jobStates,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list jobs: %s", err.Error()))
		return
	}

	ctx.JSON(200, jobs)
}
