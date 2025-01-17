// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net/http"

	_ "github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/gin-gonic/gin"
)

// UpdateProvider godoc
//
//	@Tags			provider
//	@Summary		Update provider
//	@Description	Update provider
//	@Param			updateProviderDto	body	UpdateProviderDTO	true	"Update provider"
//	@Param			runnerId			path	string				true	"Runner ID"
//	@Param			providerName		path	string				true	"Provider name"
//	@Success		200
//	@Router			/runner/{runnerId}/provider/{providerName}/update [post]
//
//	@id				UpdateProvider
func UpdateProvider(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")
	providerName := ctx.Param("providerName")

	var updateProviderMetadataDto services.UpdateProviderDTO
	err := ctx.BindJSON(&updateProviderMetadataDto)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	err = server.RunnerService.UpdateProvider(ctx.Request.Context(), runnerId, providerName, updateProviderMetadataDto)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to update provider: %w", err))
		return
	}

	ctx.Status(200)
}
