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
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)
	if _, err := server.ProviderManager.GetProvider(req.Name); err == nil {
		err := server.ProviderManager.UninstallProvider(req.Name)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to uninstall current provider: %s", err.Error()))
			return
		}
	}

	downloadPath, err := server.ProviderManager.DownloadProvider(req.DownloadUrls, req.Name, true)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to download provider: %s", err.Error()))
		return
	}

	err = server.ProviderManager.RegisterProvider(downloadPath)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to register provider: %s", err.Error()))
		return
	}

	ctx.Status(200)
}
