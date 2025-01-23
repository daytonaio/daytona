// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListRunners 			godoc
//
//	@Tags			runner
//	@Summary		List runners
//	@Description	List runners
//	@Produce		json
//	@Success		200	{array}	RunnerDTO
//	@Router			/runner [get]
//
//	@id				ListRunners
func ListRunners(ctx *gin.Context) {
	server := server.GetInstance(nil)

	runners, err := server.RunnerService.List(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to register runner: %w", err))
		return
	}

	ctx.JSON(200, runners)
}
