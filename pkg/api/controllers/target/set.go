// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// SetTarget godoc
//
//	@Tags			target
//	@Summary		Set a target
//	@Description	Set a target
//	@Param			target	body	ProviderTarget	true	"Target to set"
//	@Success		201
//	@Router			/target [put]
//
//	@id				SetTarget
func SetTarget(ctx *gin.Context) {
	var req provider.ProviderTarget
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	target, err := server.ProviderTargetService.Find(req.Name)
	if err == nil {
		target.Options = req.Options
		target.ProviderInfo = req.ProviderInfo
	} else {
		target = &req
	}

	err = server.ProviderTargetService.Save(target)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set target: %w", err))
		return
	}

	ctx.Status(201)
}
