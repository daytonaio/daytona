// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// DeleteRunner 			godoc
//
//	@Tags			runner
//	@Summary		Delete runner
//	@Description	Delete runner
//	@Param			runnerId	path	string	true	"Runner ID"
//	@Success		200
//	@Router			/runner/{runnerId} [delete]
//
//	@id				DeleteRunner
func DeleteRunner(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")

	server := server.GetInstance(nil)

	err := server.RunnerService.DeleteRunner(ctx.Request.Context(), runnerId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to delete runner: %w", err))
		return
	}

	ctx.Status(200)
}
