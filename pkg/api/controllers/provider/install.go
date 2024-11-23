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

// InstallProvider godoc
//
//	@Tags			provider
//	@Summary		Install a provider
//	@Description	Install a provider
//	@Accept			json
//	@Param			provider	body	InstallProviderRequest	true	"Provider to install"
//	@Success		200
//	@Router			/provider/install [post]
//
//	@id				InstallProvider
func InstallProvider(ctx *gin.Context) {
	var req dto.InstallProviderRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)
	if _, err := server.ProviderManager.GetProvider(req.Name); err == nil {
		err := server.ProviderManager.UninstallProvider(req.Name)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to uninstall current provider: %w", err))
			return
		}
	}

	downloadPath, err := server.ProviderManager.DownloadProvider(ctx.Request.Context(), req.DownloadUrls, req.Name)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to download provider: %w", err))
		return
	}

	err = server.ProviderManager.RegisterProvider(downloadPath, true)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to register provider: %w", err))
		return
	}

	ctx.Status(200)
}
