// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// UpdateWorkspaceLabels 			godoc
//
//	@Tags			workspace
//	@Summary		Update workspace labels
//	@Description	Update workspace labels
//	@Param			workspaceId	path		string				true	"Workspace ID or Name"
//	@Param			labels		body		map[string]string	true	"Labels"
//	@Success		200			{object}	WorkspaceDTO
//	@Router			/workspace/{workspaceId}/labels [post]
//
//	@id				UpdateWorkspaceLabels
func UpdateWorkspaceLabels(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	var req map[string]string
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)

	w, err := server.WorkspaceService.UpdateLabels(ctx.Request.Context(), workspaceId, req)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to update labels for workspace %s: %w", workspaceId, err))
		return
	}

	apiKeyType, ok := ctx.Get("apiKeyType")
	if !ok || apiKeyType == models.ApiKeyTypeClient {
		util.HideDaytonaEnvVars(&w.EnvVars)
		util.HideDaytonaEnvVars(&w.Target.EnvVars)
		w.ApiKey = ""
		w.Target.ApiKey = ""
	}

	ctx.JSON(200, w)
}
