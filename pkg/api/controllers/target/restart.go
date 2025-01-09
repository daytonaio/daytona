// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// RestartTarget 			godoc
//
//	@Tags			target
//	@Summary		Restart target
//	@Description	Restart target
//	@Param			targetId	path	string	true	"Target ID or Name"
//	@Success		200
//	@Router			/target/{targetId}/restart [post]
//
//	@id				RestartTarget
func RestartTarget(ctx *gin.Context) {
	targetId := ctx.Param("targetId")

	server := server.GetInstance(nil)

	err := server.TargetService.Restart(ctx.Request.Context(), targetId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to restart target %s: %w", targetId, err))
		return
	}

	ctx.Status(200)
}
