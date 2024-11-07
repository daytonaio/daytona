// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/api/util"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs/dto"
	"github.com/gin-gonic/gin"
)

// SetTargetConfig godoc
//
//	@Tags			target-config
//	@Summary		Set a target config
//	@Description	Set a target config
//	@Param			targetConfig	body		CreateTargetConfigDTO	true	"Target config to set"
//	@Success		200				{object}	TargetConfig
//	@Router			/target-config [put]
//
//	@id				SetTargetConfig
func SetTargetConfig(ctx *gin.Context) {
	var req dto.CreateTargetConfigDTO
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	targetConfig := conversion.ToTargetConfig(req)

	err = server.TargetConfigService.Save(targetConfig)
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
