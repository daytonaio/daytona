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

// HealthCheck      godoc
//
//	@Tags			provider
//	@Summary		Provider health check
//	@Description	Provider health check
//	@Success		200
//	@Router			/provider/health [get]
//	@id				HealthCheck
func HealthCheck(ctx *gin.Context) {
	var req dto.Provider
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	server := server.GetInstance(nil)
	exist := server.ProviderManager.IsInitialized(req.Name)
	if !exist {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to initialize provider: %w", err))
		return
	}
	ctx.Status(200)
}
