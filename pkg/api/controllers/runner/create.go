// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/gin-gonic/gin"
)

// CreateRunner 			godoc
//
//	@Tags			runner
//	@Summary		Create a runner
//	@Description	Create a runner
//	@Param			runner	body	CreateRunnerDTO	true	"Runner"
//	@Produce		json
//	@Success		200	{object}	CreateRunnerResultDTO
//	@Router			/runner [post]
//
//	@id				CreateRunner
func CreateRunner(ctx *gin.Context) {
	var createRunnerReq services.CreateRunnerDTO
	err := ctx.BindJSON(&createRunnerReq)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	r, err := server.RunnerService.CreateRunner(ctx.Request.Context(), createRunnerReq)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create runner: %w", err))
		return
	}

	ctx.JSON(200, services.CreateRunnerResultDTO{
		Runner: r.Runner,
		ApiKey: r.ApiKey,
	})
}
