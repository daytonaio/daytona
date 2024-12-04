// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/api/util"
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
//	@Param			verbose		query		bool	false	"Verbose"
//	@Success		200			{object}	WorkspaceDTO
//	@Router			/workspace/{workspaceId} [get]
//
//	@id				GetWorkspace
func GetWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")
	verboseQuery := ctx.Query("verbose")
	verbose := false
	var err error

	if verboseQuery != "" {
		verbose, err = strconv.ParseBool(verboseQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for verbose flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	w, err := server.WorkspaceService.GetWorkspace(ctx.Request.Context(), workspaceId, services.WorkspaceRetrievalParams{Verbose: verbose})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if stores.IsWorkspaceNotFound(err) || services.IsWorkspaceDeleted(err) {
			statusCode = http.StatusNotFound
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to find workspace: %w", err))
		return
	}

	util.HideDaytonaEnvVars(&w.EnvVars)
	util.HideDaytonaEnvVars(&w.Target.EnvVars)

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
//	@Param			verbose	query	bool	false	"Verbose"
//
//	@id				ListWorkspaces
func ListWorkspaces(ctx *gin.Context) {
	verboseQuery := ctx.Query("verbose")
	verbose := false
	var err error

	if verboseQuery != "" {
		verbose, err = strconv.ParseBool(verboseQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for verbose flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	workspaceList, err := server.WorkspaceService.ListWorkspaces(ctx.Request.Context(), services.WorkspaceRetrievalParams{Verbose: verbose})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list workspaces: %w", err))
		return
	}

	for i, _ := range workspaceList {
		util.HideDaytonaEnvVars(&workspaceList[i].EnvVars)
		util.HideDaytonaEnvVars(&workspaceList[i].Target.EnvVars)
	}

	ctx.JSON(200, workspaceList)
}
