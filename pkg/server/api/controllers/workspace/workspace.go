// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
	"github.com/daytonaio/daytona/pkg/server/workspaceservice"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
	workspaces, err := db.ListWorkspaces()
	verbose := ctx.Query("verbose")

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to list workspaces"))
		return
	}

	response := []dto.WorkspaceDTO{}

	for _, workspace := range workspaces {
		var workspaceInfo *types.WorkspaceInfo
		if verbose != "" {
			isVerbose, err := strconv.ParseBool(verbose)
			if err != nil {
				ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for verbose flag"))
				return
			}

			if isVerbose {
				workspaceInfo, err = provisioner.GetWorkspaceInfo(workspace)
				if err != nil {
					log.Error(fmt.Errorf("failed to get workspace info for %s", workspace.Name))
					// return
				}
			}
		}

		response = append(response, dto.WorkspaceDTO{
			Workspace: *workspace,
			Info:      workspaceInfo,
		})
	}

	ctx.JSON(200, response)
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
