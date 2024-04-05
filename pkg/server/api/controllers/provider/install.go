// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/provider/dto"
	api_util "github.com/daytonaio/daytona/pkg/server/api/util"
	"github.com/daytonaio/daytona/pkg/server/config"
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

	if _, err := manager.GetProvider(req.Name); err == nil {
		err := manager.UninstallProvider(req.Name)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to uninstall current provider: %s", err.Error()))
			return
		}
	}

	c, err := config.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %s", err.Error()))
		return
	}

	downloadPath := filepath.Join(c.ProvidersDir, req.Name, req.Name)

	err = manager.DownloadProvider(req.DownloadUrls, downloadPath)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to download provider: %s", err.Error()))
		return
	}

	logsDir, err := config.GetWorkspaceLogsDir()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get workspace logs dir: %s", err.Error()))
		return
	}

	err = manager.RegisterProvider(downloadPath, api_util.GetDaytonaScriptUrl(c), util.GetFrpcServerUrl(c), util.GetFrpcApiUrl(c), logsDir)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to register provider: %s", err.Error()))
		return
	}

	ctx.Status(200)
}
