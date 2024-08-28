// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// GenerateApiKey 			godoc
//
//	@Tags			apiKey
//	@Summary		Generate an API key
//	@Description	Generate an API key
//	@Produce		plain
//	@Param			apiKeyName	path		string	true	"API key name"
//	@Success		200			{string}	apiKey
//	@Router			/apikey/{apiKeyName} [post]
//
//	@id				GenerateApiKey
func GenerateApiKey(ctx *gin.Context) {
	apiKeyName := ctx.Param("apiKeyName")

	server := server.GetInstance(nil)

	response, err := server.ApiKeyService.Generate(apikey.ApiKeyTypeClient, apiKeyName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get API keys: %w", err))
		return
	}

	ctx.String(200, response)
}
