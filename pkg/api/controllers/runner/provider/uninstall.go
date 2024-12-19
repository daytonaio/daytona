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

// UninstallProvider godoc
//
//	@Tags			provider
//	@Summary		Uninstall provider
//	@Description	Uninstall provider
//	@Param			runnerId		path	string	true	"Runner ID"
//	@Param			providerName	path	string	true	"Provider name"
//	@Success		200
//	@Router			/runner/{runnerId}/provider/{providerName}/uninstall [post]
//
//	@id				UninstallProvider
func UninstallProvider(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")
	providerName := ctx.Param("providerName")

	server := server.GetInstance(nil)

	err := server.RunnerService.UninstallProvider(ctx.Request.Context(), runnerId, providerName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to uninstall provider: %w", err))
		return
	}

	ctx.Status(200)
}
