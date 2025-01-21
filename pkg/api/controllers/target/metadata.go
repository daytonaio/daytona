// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/controllers/target/dto"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// SetTargetMetadata 			godoc
//
//	@Tags			target
//	@Summary		Set target metadata
//	@Description	Set target metadata
//	@Param			targetId	path	string				true	"Target ID"
//	@Param			setMetadata	body	SetTargetMetadata	true	"Set Metadata"
//	@Success		200
//	@Router			/target/{targetId}/metadata [post]
//
//	@id				SetTargetMetadata
func SetTargetMetadata(ctx *gin.Context) {
	targetId := ctx.Param("targetId")

	var setTargetMetadataDTO dto.SetTargetMetadata
	err := ctx.BindJSON(&setTargetMetadataDTO)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	_, err = server.TargetService.SetTargetMetadata(ctx.Request.Context(), targetId, &models.TargetMetadata{
		Uptime: setTargetMetadataDTO.Uptime,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set target metadata for %s: %w", targetId, err))
		return
	}

	ctx.Status(200)
}
