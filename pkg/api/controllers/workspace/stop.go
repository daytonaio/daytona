// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// StopWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Stop workspace
//	@Description	Stop workspace
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Success		200
//	@Router			/workspace/{workspaceId}/stop [post]
//
//	@id				StopWorkspace
func StopWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	server := server.GetInstance(nil)

	err := server.WorkspaceService.StopWorkspace(ctx.Request.Context(), workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop workspace %s: %w", workspaceId, err))
		return
	}

	ctx.Status(200)
}
