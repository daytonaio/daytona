// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/pkg/api/controllers/target/dto"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/gin-gonic/gin"
)

// SetWorkspaceState 			godoc
//
//	@Tags			workspace
//	@Summary		Set workspace state
//	@Description	Set workspace state
//	@Param			workspaceId	path	string				true	"Workspace ID"
//	@Param			setState	body	SetWorkspaceState	true	"Set State"
//	@Success		200
//	@Router			/workspace/{workspaceId}/state [post]
//
//	@id				SetWorkspaceState
func SetWorkspaceState(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	var setWorkspaceStateDTO dto.SetWorkspaceState
	err := ctx.BindJSON(&setWorkspaceStateDTO)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	_, err = server.WorkspaceService.SetWorkspaceState(workspaceId, &workspace.WorkspaceState{
		Uptime:    setWorkspaceStateDTO.Uptime,
		UpdatedAt: time.Now().Format(time.RFC1123),
		GitStatus: setWorkspaceStateDTO.GitStatus,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set workspace state for %s: %w", workspaceId, err))
		return
	}

	ctx.Status(200)
}
