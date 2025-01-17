// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// FindWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Find workspace
//	@Description	Find workspace
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Success		200			{object}	WorkspaceDTO
//	@Router			/workspace/{workspaceId} [get]
//
//	@id				FindWorkspace
func FindWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")
	server := server.GetInstance(nil)

	w, err := server.WorkspaceService.Find(ctx.Request.Context(), workspaceId, services.WorkspaceRetrievalParams{})
	if err != nil {
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
		w.ApiKey = ""
		w.Target.ApiKey = ""
	}

	ctx.JSON(200, w)
}

// ListWorkspaces 			godoc
//
//	@Tags			workspace
//	@Summary		List workspaces
//	@Description	List workspaces
//	@Param			labels	query	string	false	"JSON encoded labels"
//	@Produce		json
//	@Success		200	{array}	WorkspaceDTO
//	@Router			/workspace [get]
//
//	@id				ListWorkspaces
func ListWorkspaces(ctx *gin.Context) {
	labelsQuery := ctx.Query("labels")

	var labels map[string]string

	if labelsQuery != "" {
		err := json.Unmarshal([]byte(labelsQuery), &labels)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid filters: %w", err))
			return
		}
	}

	server := server.GetInstance(nil)

	workspaceList, err := server.WorkspaceService.List(ctx.Request.Context(), services.WorkspaceRetrievalParams{
		Labels: labels,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list workspaces: %w", err))
		return
	}

	apiKeyType, ok := ctx.Get("apiKeyType")
	if !ok || apiKeyType == models.ApiKeyTypeClient {
		for i := range workspaceList {
			util.HideDaytonaEnvVars(&workspaceList[i].EnvVars)
			util.HideDaytonaEnvVars(&workspaceList[i].Target.EnvVars)
			workspaceList[i].ApiKey = ""
			workspaceList[i].Target.ApiKey = ""
		}
	}

	ctx.JSON(200, workspaceList)
}
