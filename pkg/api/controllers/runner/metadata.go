// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/controllers/runner/dto"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// UpdateRunnerMetadata 			godoc
//
//	@Tags			runner
//	@Summary		Update runner metadata
//	@Description	Update runner metadata
//	@Param			runnerId		path	string					true	"Runner ID"
//	@Param			runnerMetadata	body	UpdateRunnerMetadataDTO	true	"Runner Metadata"
//	@Success		200
//	@Router			/runner/{runnerId}/metadata [post]
//
//	@id				UpdateRunnerMetadata
func UpdateRunnerMetadata(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")

	var runnerMetadata dto.UpdateRunnerMetadataDTO
	err := ctx.BindJSON(&runnerMetadata)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	err = server.RunnerService.UpdateRunnerMetadata(ctx.Request.Context(), runnerId, &models.RunnerMetadata{
		RunnerId:    runnerId,
		Uptime:      runnerMetadata.Uptime,
		Providers:   runnerMetadata.Providers,
		RunningJobs: runnerMetadata.RunningJobs,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set runner metadata for %s: %w", runnerId, err))
		return
	}

	ctx.Status(200)
}
