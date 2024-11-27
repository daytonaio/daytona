// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// SetWorkspaceMetadata 			godoc
//
//	@Tags			workspace
//	@Summary		Set workspace metadata
//	@Description	Set workspace metadata
//	@Param			workspaceId	path	string					true	"Workspace ID"
//	@Param			setMetadata	body	SetWorkspaceMetadata	true	"Set Metadata"
//	@Success		200
//	@Router			/workspace/{workspaceId}/metadata [post]
//
//	@id				SetWorkspaceMetadata
func SetWorkspaceMetadata(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	var setWorkspaceMetadataDTO dto.SetWorkspaceMetadata
	err := ctx.BindJSON(&setWorkspaceMetadataDTO)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	_, err = server.WorkspaceService.SetWorkspaceMetadata(workspaceId, &models.WorkspaceMetadata{
		Uptime:    setWorkspaceMetadataDTO.Uptime,
		GitStatus: setWorkspaceMetadataDTO.GitStatus,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set workspace metadata for %s: %w", workspaceId, err))
		return
	}

	ctx.Status(200)
}
