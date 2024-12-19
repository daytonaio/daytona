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

// RegisterRunner 			godoc
//
//	@Tags			runner
//	@Summary		Register a runner
//	@Description	Register a runner
//	@Param			runner	body	RegisterRunnerDTO	true	"Register runner"
//	@Produce		json
//	@Success		200	{object}	RegisterRunnerResultDTO
//	@Router			/runner [post]
//
//	@id				RegisterRunner
func RegisterRunner(ctx *gin.Context) {
	var registerRunnerReq services.RegisterRunnerDTO
	err := ctx.BindJSON(&registerRunnerReq)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	r, err := server.RunnerService.RegisterRunner(ctx.Request.Context(), registerRunnerReq)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to register runner: %w", err))
		return
	}

	ctx.JSON(200, services.RegisterRunnerResultDTO{
		Runner: r.Runner,
		ApiKey: r.ApiKey,
	})
}
