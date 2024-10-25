// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
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
	var createTargetReq dto.CreateTargetDTO
	err := ctx.BindJSON(&createTargetReq)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	w, err := server.TargetService.CreateTarget(ctx.Request.Context(), createTargetReq)
	if err != nil {
		if errors.Is(err, targets.ErrTargetAlreadyExists) {
			ctx.AbortWithError(http.StatusConflict, fmt.Errorf("target already exists: %w", err))
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create target: %w", err))
		return
	}

	ctx.JSON(200, w)
}
