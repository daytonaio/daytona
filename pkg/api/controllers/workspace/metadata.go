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

// UpdateWorkspaceMetadata 			godoc
//
//	@Tags			workspace
//	@Summary		Update workspace metadata
//	@Description	Update workspace metadata
//	@Param			workspaceId			path	string						true	"Workspace ID"
//	@Param			workspaceMetadata	body	UpdateWorkspaceMetadataDTO	true	"Workspace Metadata"
//	@Success		200
//	@Router			/workspace/{workspaceId}/metadata [post]
//
//	@id				UpdateWorkspaceMetadata
func UpdateWorkspaceMetadata(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	var updateDTO dto.UpdateWorkspaceMetadataDTO
	err := ctx.BindJSON(&updateDTO)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	_, err = server.WorkspaceService.UpdateWorkspaceMetadata(ctx.Request.Context(), workspaceId, &models.WorkspaceMetadata{
		Uptime:    updateDTO.Uptime,
		GitStatus: updateDTO.GitStatus,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set workspace metadata for %s: %w", workspaceId, err))
		return
	}

	ctx.Status(200)
}

// UpdateWorkspaceProviderMetadata 			godoc
//
//	@Tags			workspace
//	@Summary		Update workspace provider metadata
//	@Description	Update workspace provider metadata
//	@Param			workspaceId	path	string								true	"Workspace ID"
//	@Param			metadata	body	UpdateWorkspaceProviderMetadataDTO	true	"Provider metadata"
//	@Success		200
//	@Router			/workspace/{workspaceId}/provider-metadata [post]
//
//	@id				UpdateWorkspaceProviderMetadata
func UpdateWorkspaceProviderMetadata(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	var metadata dto.UpdateWorkspaceProviderMetadataDTO
	err := ctx.BindJSON(&metadata)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	err = server.WorkspaceService.UpdateWorkspaceProviderMetadata(ctx.Request.Context(), workspaceId, metadata.Metadata)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to update workspace provider metadata for %s: %w", workspaceId, err))
		return
	}

	ctx.Status(200)
}
