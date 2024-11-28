// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// RemoveTargetConfig godoc
//
//	@Tags			target-config
//	@Summary		Remove a target config
//	@Description	Remove a target config
//	@Param			configId	path	string	true	"Target Config Id"
//	@Success		204
//	@Router			/target-config/{configId} [delete]
//
//	@id				RemoveTargetConfig
func RemoveTargetConfig(ctx *gin.Context) {
	configId := ctx.Param("configId")

	server := server.GetInstance(nil)

	err := server.TargetConfigService.Delete(configId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove target config: %w", err))
		return
	}

	ctx.Status(204)
}
