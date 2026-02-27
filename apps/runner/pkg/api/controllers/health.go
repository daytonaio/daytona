// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0
package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/daytonaio/runner/internal"
	"github.com/daytonaio/runner/pkg/runner"
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
	pingCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
	defer cancel()

	runner, err := runner.GetInstance(nil)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = runner.Docker.Ping(pingCtx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": internal.Version,
	})
}
