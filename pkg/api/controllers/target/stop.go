// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// StopTarget 			godoc
//
//	@Tags			target
//	@Summary		Stop target
//	@Description	Stop target
//	@Param			targetId	path	string	true	"Target ID or Name"
//	@Success		200
//	@Router			/target/{targetId}/stop [post]
//
//	@id				StopTarget
func StopTarget(ctx *gin.Context) {
	targetId := ctx.Param("targetId")

	server := server.GetInstance(nil)

	err := server.TargetService.Stop(ctx.Request.Context(), targetId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop target %s: %w", targetId, err))
		return
	}

	ctx.Status(200)
}
