// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/controllers/provider/dto"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListProviders godoc
//
//	@Tags			provider
//	@Summary		List providers
//	@Description	List providers
//	@Produce		json
//	@Success		200	{array}	dto.Provider
//	@Router			/provider [get]
//
//	@id				ListProviders
func ListProviders(ctx *gin.Context) {
	server := server.GetInstance(nil)
	providers := server.ProviderManager.GetProviders()

	result := []dto.Provider{}
	for _, provider := range providers {
		info, err := provider.GetInfo()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get provider: %w", err))
			return
		}

		result = append(result, dto.Provider{
			Name:    info.Name,
			Version: info.Version,
		})
	}

	ctx.JSON(200, result)
}
