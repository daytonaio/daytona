// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// RestartWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Restart workspace
//	@Description	Restart workspace
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Success		200
//	@Router			/workspace/{workspaceId}/restart [post]
//
//	@id				RestartWorkspace
func RestartWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	server := server.GetInstance(nil)

	err := server.WorkspaceService.Restart(ctx.Request.Context(), workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to restart workspace %s: %w", workspaceId, err))
		return
	}

	ctx.Status(200)
}
