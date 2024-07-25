// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/gin-gonic/gin"
)

// CreateWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Create a workspace
//	@Description	Create a workspace
//	@Param			workspace	body	CreateWorkspaceRequest	true	"Create workspace"
//	@Produce		json
//	@Success		200	{object}	Workspace
//	@Router			/workspace [post]
//
//	@id				CreateWorkspace
func CreateWorkspace(ctx *gin.Context) {
	var createWorkspaceReq dto.CreateWorkspaceRequest
	err := ctx.BindJSON(&createWorkspaceReq)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	w, err := server.WorkspaceService.CreateWorkspace(ctx.Request.Context(), createWorkspaceReq)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create workspace: %s", err.Error()))
		return
	}

	ctx.JSON(200, w)
}
