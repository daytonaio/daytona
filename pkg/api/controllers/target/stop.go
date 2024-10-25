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

// StopProject 			godoc
//
//	@Tags			target
//	@Summary		Stop project
//	@Description	Stop project
//	@Param			targetId	path	string	true	"Target ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Success		200
//	@Router			/target/{targetId}/{projectId}/stop [post]
//
//	@id				StopProject
func StopProject(ctx *gin.Context) {
	targetId := ctx.Param("targetId")
	projectId := ctx.Param("projectId")

	server := server.GetInstance(nil)

	err := server.TargetService.StopProject(ctx.Request.Context(), targetId, projectId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop project %s: %w", projectId, err))
		return
	}

	ctx.Status(200)
}
