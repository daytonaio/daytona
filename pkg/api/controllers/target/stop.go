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

	err := server.TargetService.StopTarget(ctx.Request.Context(), targetId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop target %s: %w", targetId, err))
		return
	}

	ctx.Status(200)
}

// StopWorkspace 			godoc
//
//	@Tags			target
//	@Summary		Stop workspace
//	@Description	Stop workspace
//	@Param			targetId	path	string	true	"Target ID or Name"
//	@Param			workspaceId	path	string	true	"Workspace ID"
//	@Success		200
//	@Router			/target/{targetId}/{workspaceId}/stop [post]
//
//	@id				StopWorkspace
func StopWorkspace(ctx *gin.Context) {
	targetId := ctx.Param("targetId")
	workspaceId := ctx.Param("workspaceId")

	server := server.GetInstance(nil)

	err := server.TargetService.StopWorkspace(ctx.Request.Context(), targetId, workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop workspace %s: %w", workspaceId, err))
		return
	}

	ctx.Status(200)
}
