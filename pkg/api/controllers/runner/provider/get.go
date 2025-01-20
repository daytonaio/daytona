// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// GetRunnerProviders 			godoc
//
//	@Tags			provider
//	@Summary		Get runner providers
//	@Description	Get runner providers
//	@Param			runnerId	path	string	true	"Runner ID"
//	@Produce		json
//	@Success		200	{array}	ProviderInfo
//	@Router			/runner/{runnerId}/provider [get]
//
//	@id				GetRunnerProviders
func GetRunnerProviders(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")

	server := server.GetInstance(nil)

	r, err := server.RunnerService.Find(ctx.Request.Context(), runnerId)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if stores.IsRunnerNotFound(err) {
			statusCode = http.StatusNotFound
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to get runner: %w", err))
		return
	}

	ctx.JSON(200, r.Metadata.Providers)
}
