// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// ListEnvironmentVariables godoc
//
//	@Tags			envVar
//	@Summary		List environment variables
//	@Description	List environment variables
//	@Produce		json
//	@Success		200	{array}	models.EnvironmentVariable
//	@Router			/env [get]
//
//	@id				ListEnvironmentVariables
func ListEnvironmentVariables(ctx *gin.Context) {
	server := server.GetInstance(nil)
	envVars, err := server.EnvironmentVariableService.List()
	if err != nil {
		if stores.IsEnvironmentVariableNotFound(err) {
			ctx.JSON(200, []*models.EnvironmentVariable{})
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list environment variables: %w", err))
		return
	}

	ctx.JSON(200, envVars)
}

// SetEnvironmentVariable godoc
//
//	@Tags			envVar
//	@Summary		Set environment variable
//	@Description	Set environment variable
//	@Accept			json
//	@Param			environmentVariable	body	models.EnvironmentVariable	true	"Environment Variable"
//	@Success		201
//	@Router			/env [put]
//
//	@id				SetEnvironmentVariable
func SetEnvironmentVariable(ctx *gin.Context) {
	var req models.EnvironmentVariable
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	server := server.GetInstance(nil)
	err = server.EnvironmentVariableService.Save(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save environment variable: %w", err))
		return
	}

	ctx.Status(201)
}

// DeleteEnvironmentVariable godoc
//
//	@Tags			envVar
//	@Summary		Delete environment variable
//	@Description	Delete environment variable
//	@Param			key	path	string	true	"Environment Variable Key"
//	@Success		204
//	@Router			/env/{key} [delete]
//
//	@id				DeleteEnvironmentVariable
func DeleteEnvironmentVariable(ctx *gin.Context) {
	envVarKey := ctx.Param("key")

	server := server.GetInstance(nil)

	err := server.EnvironmentVariableService.Delete(envVarKey)
	if err != nil {
		if stores.IsEnvironmentVariableNotFound(err) {
			ctx.Status(204)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove environment variable: %w", err))
		return
	}

	ctx.Status(204)
}
