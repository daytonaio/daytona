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
//	@Produce		json
//	@Success		200	{array}	TargetConfig
//	@Router			/target-config [get]
//
//	@id				ListTargetConfigs
func ListTargetConfigs(ctx *gin.Context) {
	server := server.GetInstance(nil)

	targetConfigs, err := server.TargetConfigService.List(nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list target configs: %w", err))
		return
	}

	ctx.JSON(200, targetConfigs)
}
