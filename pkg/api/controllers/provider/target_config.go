// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// GetTargetConfigManifest godoc
//
//	@Tags			provider
//	@Summary		Get provider target config manifest
//	@Description	Get provider target config manifest
//	@Param			provider	path	string	true	"Provider name"
//	@Success		200
//	@Success		200	{object}	TargetConfigManifest
//	@Router			/provider/{provider}/target-config-manifest [get]
//
//	@id				GetTargetConfigManifest
func GetTargetConfigManifest(ctx *gin.Context) {
	providerName := ctx.Param("provider")

	server := server.GetInstance(nil)

	p, err := server.ProviderManager.GetProvider(providerName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("provider not found: %w", err))
		return
	}

	manifest, err := (*p).GetTargetConfigManifest()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get provider manifest: %w", err))
		return
	}

	ctx.JSON(200, manifest)
}
