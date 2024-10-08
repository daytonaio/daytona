// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"net/http"

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
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
