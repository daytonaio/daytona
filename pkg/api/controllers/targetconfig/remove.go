// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// RemoveTargetConfig godoc
//
//	@Tags			target-config
//	@Summary		Remove a target config
//	@Description	Remove a target config
//	@Param			configName	path	string	true	"Target Config name"
//	@Success		204
//	@Router			/target-config/{configName} [delete]
//
//	@id				RemoveTargetConfig
func RemoveTargetConfig(ctx *gin.Context) {
	configName := ctx.Param("configName")

	server := server.GetInstance(nil)

	targetConfig, err := server.TargetConfigService.Find(&provider.TargetConfigFilter{
		Name: &configName,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to find target config: %w", err))
		return
	}

	err = server.TargetConfigService.Delete(targetConfig)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove target config: %w", err))
		return
	}

	ctx.Status(204)
}
