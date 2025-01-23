// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// InstallProvider godoc
//
//	@Tags			provider
//	@Summary		Install provider
//	@Description	Install provider
//	@Param			runnerId		path	string	true	"Runner ID"
//	@Param			providerName	path	string	true	"Provider name"
//	@Param			providerVersion	query	string	false	"Provider version - defaults to 'latest'"
//	@Success		200
//	@Router			/runner/{runnerId}/provider/{providerName}/install [post]
//
//	@id				InstallProvider
func InstallProvider(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")
	providerName := ctx.Param("providerName")
	providerVersion := ctx.DefaultQuery("providerVersion", "latest")

	config, err := server.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %w", err))
		return
	}

	server := server.GetInstance(nil)

	installedProviders, err := server.RunnerService.ListProviders(ctx, &runnerId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to fetch installed providers: %w", err))
		return
	}

	for _, provider := range installedProviders {
		if provider.Name == providerName {
			ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("provider %s is already installed", providerName))
			return
		}
	}

	err = server.RunnerService.InstallProvider(ctx.Request.Context(), runnerId, providerName, providerVersion, config.RegistryUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to install provider: %w", err))
		return
	}

	ctx.Status(200)
}
