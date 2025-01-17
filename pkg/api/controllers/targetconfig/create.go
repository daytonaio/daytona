// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/gin-gonic/gin"
)

// CreateTargetConfig godoc
//
//	@Tags			target-config
//	@Summary		Create a target config
//	@Description	Create a target config
//	@Param			targetConfig	body		CreateTargetConfigDTO	true	"Target config to create"
//	@Param			showOptions		query		bool					false	"Show target config options"
//	@Success		200				{object}	TargetConfig
//	@Router			/target-config [post]
//
//	@id				CreateTargetConfig
func CreateTargetConfig(ctx *gin.Context) {
	showTargetConfigOptionsQuery := ctx.Query("showOptions")
	var showTargetConfigOptions bool
	if showTargetConfigOptionsQuery == "true" {
		showTargetConfigOptions = true
	}

	var req services.CreateTargetConfigDTO
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	targetConfig, err := server.TargetConfigService.Create(ctx.Request.Context(), req)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set target config: %w", err))
		return
	}

	if !showTargetConfigOptions {
		targetConfig.Options = ""
	}

	ctx.JSON(200, targetConfig)
}
