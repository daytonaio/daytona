// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net/http"

	_ "github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListProviders godoc
//
//	@Tags			provider
//	@Summary		List providers
//	@Description	List providers
//	@Param			runnerId	query	string	false	"Runner ID"
//	@Produce		json
//	@Success		200	{array}	models.ProviderInfo
//	@Router			/runner/provider [get]
//
//	@id				ListProviders
func ListProviders(ctx *gin.Context) {
	runnerIdQuery := ctx.Query("runnerId")

	var runnerId *string
	if runnerIdQuery != "" {
		runnerId = &runnerIdQuery
	}

	server := server.GetInstance(nil)
	providers, err := server.RunnerService.ListProviders(ctx.Request.Context(), runnerId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list providers: %w", err))
		return
	}

	ctx.JSON(200, providers)
}
