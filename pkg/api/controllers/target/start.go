// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// StartTarget 			godoc
//
//	@Tags			target
//	@Summary		Start target
//	@Description	Start target
//	@Param			targetId	path	string	true	"Target ID or Name"
//	@Success		200
//	@Router			/target/{targetId}/start [post]
//
//	@id				StartTarget
func StartTarget(ctx *gin.Context) {
	targetId := ctx.Param("targetId")

	server := server.GetInstance(nil)

	err := server.TargetService.StartTarget(ctx.Request.Context(), targetId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to start target %s: %w", targetId, err))
		return
	}

	ctx.Status(200)
}
