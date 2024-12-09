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
//	@Produce		json
//	@Success		200	{array}	models.ProviderInfo
//	@Router			/provider [get]
//
//	@id				ListProviders
func ListProviders(ctx *gin.Context) {
	server := server.GetInstance(nil)
	providers, err := server.RunnerService.ListProviders(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list providers: %w", err))
		return
	}

	ctx.JSON(200, providers)
}
