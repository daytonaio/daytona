// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListTargets godoc
//
//	@Tags			target
//	@Summary		List targets
//	@Description	List targets
//	@Produce		json
//	@Success		200	{array}	ProviderTarget
//	@Router			/target [get]
//
//	@id				ListTargets
func ListTargets(ctx *gin.Context) {
	server := server.GetInstance(nil)

	targets, err := server.ProviderTargetService.List()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list targets: %s", err.Error()))
		return
	}

	ctx.JSON(200, targets)
}
