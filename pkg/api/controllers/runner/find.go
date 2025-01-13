// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// FindRunner 			godoc
//
//	@Tags			runner
//	@Summary		Find a runner
//	@Description	Find a runner
//	@Param			runnerId	path	string	true	"Runner ID"
//	@Produce		json
//	@Success		200	{object}	RunnerDTO
//	@Router			/runner/{runnerId} [get]
//
//	@id				FindRunner
func FindRunner(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")

	server := server.GetInstance(nil)

	r, err := server.RunnerService.FindRunner(ctx.Request.Context(), runnerId)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if stores.IsRunnerNotFound(err) {
			statusCode = http.StatusNotFound
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to get runner: %w", err))
		return
	}

	ctx.JSON(200, r)
}
