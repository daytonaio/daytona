// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListClientApiKeys 			godoc
//
//	@Tags			apiKey
//	@Summary		List API keys
//	@Description	List API keys
//	@Produce		json
//	@Success		200	{array}	ApiKey
//	@Router			/apikey [get]
//
//	@id				ListClientApiKeys
func ListClientApiKeys(ctx *gin.Context) {
	server := server.GetInstance(nil)

	response, err := server.ApiKeyService.ListClientKeys(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get client API keys: %w", err))
		return
	}

	ctx.JSON(200, response)
}

// RevokeApiKey		godoc
//
//	@Tags			apiKey
//	@Summary		Revoke API key
//	@Description	Revoke API key
//	@Param			apiKeyName	path	string	true	"API key name"
//	@Success		200
//	@Router			/apikey/{apiKeyName} [delete]
//
//	@id				RevokeApiKey
func RevokeApiKey(ctx *gin.Context) {
	apiKeyName := ctx.Param("apiKeyName")

	server := server.GetInstance(nil)

	err := server.ApiKeyService.Revoke(ctx.Request.Context(), apiKeyName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to revoke api key: %w", err))
		return
	}

	ctx.Status(200)
}
