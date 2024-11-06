// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// SetDefaultTarget godoc
//
//	@Tags			target
//	@Summary		Set target to be used by default
//	@Description	Set target to be used by default
//	@Param			targetId	path	string	true	"Target ID or name"
//	@Success		200
//	@Router			/target/{targetId}/set-default [patch]
//
//	@id				SetDefaultTarget
func SetDefaultTarget(ctx *gin.Context) {
	targetId := ctx.Param("targetId")

	server := server.GetInstance(nil)

	err := server.TargetService.SetDefault(ctx.Request.Context(), targetId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to set target to default: %s", err.Error()))
		return
	}

	ctx.Status(200)
}
