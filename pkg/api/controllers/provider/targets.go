// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// GetTargetManifest godoc
//
//	@Tags			provider
//	@Summary		Get provider target manifest
//	@Description	Get provider target manifest
//	@Param			provider	path	string	true	"Provider name"
//	@Success		200
//	@Success		200	{object}	ProviderTargetManifest
//	@Router			/provider/{provider}/target-manifest [get]
//
//	@id				GetTargetManifest
func GetTargetManifest(ctx *gin.Context) {
	providerName := ctx.Param("provider")

	server := server.GetInstance(nil)

	p, err := server.ProviderManager.GetProvider(providerName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("provider not found: %s", err.Error()))
		return
	}

	manifest, err := (*p).GetTargetManifest()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get provider manifest: %s", err.Error()))
		return
	}

	ctx.JSON(200, manifest)
}
