// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package job

import (
	"fmt"
	"net/http"

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
//	@Param			states			query	[]string	false	"Job States"	collectionFormat(multi)
//	@Param			actions			query	[]string	false	"Job Actions"	collectionFormat(multi)
//	@Param			resourceId		query	string		false	"Resource ID"
//	@Param			resourceType	query	string		false	"Resource Type"
//	@Produce		json
//	@Success		200	{array}	Job
//	@Router			/job [get]
//
//	@id				ListJobs
func ListJobs(ctx *gin.Context) {
	states := ctx.QueryArray("states")
	var jobStates *[]models.JobState
	if len(states) > 0 {
		jobStates = &[]models.JobState{}
		for _, s := range states {
			*jobStates = append(*jobStates, models.JobState(s))
		}
	}

	resourceIdQuery := ctx.Query("resourceId")
	var resourceId *string
	if resourceIdQuery != "" {
		resourceId = &resourceIdQuery
	}

	resourceTypeQuery := ctx.Query("resourceType")
	var resourceType *models.ResourceType
	if resourceTypeQuery != "" {
		resourceType = (*models.ResourceType)(&resourceTypeQuery)
	}

	actions := ctx.QueryArray("actions")
	var jobActions *[]models.JobAction
	if len(actions) > 0 {
		jobActions = &[]models.JobAction{}
		for _, a := range actions {
			*jobActions = append(*jobActions, models.JobAction(a))
		}
	}

	server := server.GetInstance(nil)

	jobs, err := server.JobService.List(ctx.Request.Context(), &stores.JobFilter{
		States:       jobStates,
		ResourceId:   resourceId,
		ResourceType: resourceType,
		Actions:      jobActions,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list jobs: %s", err.Error()))
		return
	}

	ctx.JSON(200, jobs)
}
