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

// RemoveTarget godoc
//
//	@Tags			target
//	@Summary		Remove a target
//	@Description	Remove a target
//	@Param			target	path	string	true	"Target name"
//	@Success		204
//	@Router			/target/{target} [delete]
//
//	@id				RemoveTarget
func RemoveTarget(ctx *gin.Context) {
	targetName := ctx.Param("target")

	server := server.GetInstance(nil)

	target, err := server.ProviderTargetService.Find(&provider.TargetFilter{
		Name: &targetName,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to find target: %w", err))
		return
	}

	err = server.ProviderTargetService.Delete(target)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove target: %w", err))
		return
	}

	ctx.Status(204)
}
