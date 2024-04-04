// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/server/workspaceservice"
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

	w, err := workspaceservice.GetWorkspace(workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get workspace: %s", err.Error()))
		return
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

	workspaceList, err := workspaceservice.ListWorkspaces(verbose)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list workspaces: %s", err.Error()))
		return
	}

	ctx.JSON(200, workspaceList)
}

// RemoveWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Remove workspace
//	@Description	Remove workspace
//	@Param			workspaceId	path	string	true	"Workspace ID"
//	@Success		200
//	@Router			/workspace/{workspaceId} [delete]
//
//	@id				RemoveWorkspace
func RemoveWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	err := workspaceservice.RemoveWorkspace(workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove workspace: %s", err.Error()))
		return
	}

	ctx.Status(200)
}
