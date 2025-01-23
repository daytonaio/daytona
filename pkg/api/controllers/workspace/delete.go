// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// DeleteWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Delete workspace
//	@Description	Delete workspace
//	@Param			workspaceId	path	string	true	"Workspace ID"
//	@Param			force		query	bool	false	"Force"
//	@Success		200
//	@Router			/workspace/{workspaceId} [delete]
//
//	@id				DeleteWorkspace
func DeleteWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")
	forceQuery := ctx.Query("force")
	var err error
	force := false

	if forceQuery != "" {
		force, err = strconv.ParseBool(forceQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for force flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	if force {
		err = server.WorkspaceService.ForceDelete(ctx.Request.Context(), workspaceId)
	} else {
		err = server.WorkspaceService.Delete(ctx.Request.Context(), workspaceId)
	}

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to delete workspace: %w", err))
		return
	}

	ctx.Status(200)
}
