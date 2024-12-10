// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// HealthCheck 			godoc
//
//	@Summary		Health check
//	@Description	Health check
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
//
//	@id				HealthCheck

func HealthCheck(ctx *gin.Context) {
	cfg, err := server.GetConfig()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	services := []uint32{cfg.HeadscalePort, cfg.LocalBuilderRegistryPort}
	for _, port := range services {
		if _, err := http.Get(fmt.Sprintf("http://localhost:%d", port)); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
			return
		}
	}

	if !server.AllProviderRegistered {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
