// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/gin-gonic/gin"
)

// InstallProvider godoc
//
//	@Tags			provider
//	@Summary		Install provider
//	@Description	Install provider
//	@Param			installProviderDto	body	InstallProviderDTO	true	"Install provider"
//	@Param			runnerId			path	string				true	"Runner ID"
//	@Success		200
//	@Router			/runner/{runnerId}/provider/install [post]
//
//	@id				InstallProvider
func InstallProvider(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")

	var installProviderMetadataDto services.InstallProviderDTO
	err := ctx.BindJSON(&installProviderMetadataDto)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)

	err = server.RunnerService.InstallProvider(ctx.Request.Context(), runnerId, installProviderMetadataDto)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to install provider: %w", err))
		return
	}

	ctx.Status(200)
}
