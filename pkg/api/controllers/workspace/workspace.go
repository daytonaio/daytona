// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// GetWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Get workspace info
//	@Description	Get workspace info
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Success		200			{object}	WorkspaceDTO
//	@Router			/workspace/{workspaceId} [get]
//
//	@id				GetWorkspace
func GetWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")
	server := server.GetInstance(nil)

	w, err := server.WorkspaceService.GetWorkspace(ctx.Request.Context(), workspaceId, services.WorkspaceRetrievalParams{})
	if err != nil {

		fmt.Println(err)
		fmt.Println(err)
		fmt.Println(err)
		fmt.Println(err)
		fmt.Println(err)

		statusCode := http.StatusInternalServerError
		if stores.IsWorkspaceNotFound(err) || services.IsWorkspaceDeleted(err) {
			statusCode = http.StatusNotFound
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to find workspace: %w", err))
		return
	}

	apiKeyType, ok := ctx.Get("apiKeyType")
	if !ok || apiKeyType == models.ApiKeyTypeClient {
		util.HideDaytonaEnvVars(&w.EnvVars)
		util.HideDaytonaEnvVars(&w.Target.EnvVars)
		w.ApiKey = "<HIDDEN>"
		w.Target.ApiKey = "<HIDDEN>"
	}

	ctx.JSON(200, w)
}

// ListWorkspaces 			godoc
//
//	@Tags			workspace
//	@Summary		List workspaces
//	@Description	List workspaces
//	@Produce		json
//	@Success		200	{array}	WorkspaceDTO
//	@Router			/workspace [get]
//
//	@id				ListWorkspaces
func ListWorkspaces(ctx *gin.Context) {
	server := server.GetInstance(nil)

	workspaceList, err := server.WorkspaceService.ListWorkspaces(ctx.Request.Context(), services.WorkspaceRetrievalParams{})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list workspaces: %w", err))
		return
	}

	apiKeyType, ok := ctx.Get("apiKeyType")
	if !ok || apiKeyType == models.ApiKeyTypeClient {
		for i, _ := range workspaceList {
			util.HideDaytonaEnvVars(&workspaceList[i].EnvVars)
			util.HideDaytonaEnvVars(&workspaceList[i].Target.EnvVars)
			workspaceList[i].ApiKey = "<HIDDEN>"
			workspaceList[i].Target.ApiKey = "<HIDDEN>"
		}
	}

	ctx.JSON(200, workspaceList)
}
