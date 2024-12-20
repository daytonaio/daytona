// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// RemoveRunner 			godoc
//
//	@Tags			runner
//	@Summary		Remove runner
//	@Description	Remove runner
//	@Param			runnerId	path	string	true	"Runner ID"
//	@Success		200
//	@Router			/runner/{runnerId} [delete]
//
//	@id				RemoveRunner
func RemoveRunner(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")

	server := server.GetInstance(nil)

	err := server.RunnerService.RemoveRunner(ctx.Request.Context(), runnerId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove runner: %w", err))
		return
	}

	ctx.Status(200)
}
