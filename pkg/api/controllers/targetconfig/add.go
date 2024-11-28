// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/util"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/gin-gonic/gin"
)

// AddTargetConfig godoc
//
//	@Tags			target-config
//	@Summary		Add a target config
//	@Description	Add a target config
//	@Param			targetConfig	body		AddTargetConfigDTO	true	"Target config to add"
//	@Success		200				{object}	TargetConfig
//	@Router			/target-config [put]
//
//	@id				AddTargetConfig
func AddTargetConfig(ctx *gin.Context) {
	var req services.AddTargetConfigDTO
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	targetConfig, err := server.TargetConfigService.Add(req)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set target config: %w", err))
		return
	}

	maskedOptions, err := util.GetMaskedOptions(server, targetConfig.ProviderInfo.Name, targetConfig.Options)
	if err != nil {
		targetConfig.Options = fmt.Sprintf("Error: %s", err.Error())
	} else {
		targetConfig.Options = maskedOptions
	}

	ctx.JSON(200, targetConfig)
}
