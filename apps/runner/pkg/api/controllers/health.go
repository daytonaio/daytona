// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0
package controllers

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
//	@Router			/ [get]
//
//	@id				HealthCheck
func HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "0.0.1",
	})
}
