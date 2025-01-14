// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/gin-gonic/gin"
)

// CreateTarget 			godoc
//
//	@Tags			target
//	@Summary		Create a target
//	@Description	Create a target
//	@Param			target	body	CreateTargetDTO	true	"Create target"
//	@Produce		json
//	@Success		200	{object}	Target
//	@Router			/target [post]
//
//	@id				CreateTarget
func CreateTarget(ctx *gin.Context) {
	var createTargetReq services.CreateTargetDTO
	err := ctx.BindJSON(&createTargetReq)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	t, err := server.TargetService.Create(ctx.Request.Context(), createTargetReq)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create target: %w", err))
		return
	}

	t.TargetConfig.Options = ""
	ctx.JSON(200, t)
}

// HandleSuccessfulCreation godoc
//
//	@Tags			target
//	@Summary		Handles successful creation of the target
//	@Description	Handles successful creation of the target
//	@Param			targetId	path	string	true	"Target ID or name"
//	@Success		200
//	@Router			/target/{targetId}/handle-successful-creation [post]
//
//	@id				HandleSuccessfulCreation
func HandleSuccessfulCreation(ctx *gin.Context) {
	targetId := ctx.Param("targetId")

	server := server.GetInstance(nil)

	err := server.TargetService.HandleSuccessfulCreation(ctx.Request.Context(), targetId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to handle successful creation of target: %s", err.Error()))
		return
	}

	ctx.Status(200)
}
