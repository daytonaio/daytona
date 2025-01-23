// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListTargetConfigs godoc
//
//	@Tags			target-config
//	@Summary		List target configs
//	@Description	List target configs
//	@Param			showOptions	query	bool	false	"Show target config options"
//	@Produce		json
//	@Success		200	{array}	TargetConfig
//	@Router			/target-config [get]
//
//	@id				ListTargetConfigs
func ListTargetConfigs(ctx *gin.Context) {
	server := server.GetInstance(nil)
	showTargetConfigOptionsQuery := ctx.Query("showOptions")
	var showTargetConfigOptions bool
	if showTargetConfigOptionsQuery == "true" {
		showTargetConfigOptions = true
	}

	targetConfigs, err := server.TargetConfigService.List(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list target configs: %w", err))
		return
	}

	if !showTargetConfigOptions {
		for _, targetConfig := range targetConfigs {
			targetConfig.Options = ""
		}
	}

	ctx.JSON(200, targetConfigs)
}
