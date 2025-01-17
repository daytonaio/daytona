// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// CreateApiKey 			godoc
//
//	@Tags			apiKey
//	@Summary		Create an API key
//	@Description	Create an API key
//	@Produce		plain
//	@Param			apiKeyName	path		string	true	"API key name"
//	@Success		200			{string}	apiKey
//	@Router			/apikey/{apiKeyName} [post]
//
//	@id				CreateApiKey
func CreateApiKey(ctx *gin.Context) {
	apiKeyName := ctx.Param("apiKeyName")

	server := server.GetInstance(nil)

	response, err := server.ApiKeyService.Create(ctx.Request.Context(), models.ApiKeyTypeClient, apiKeyName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get API keys: %w", err))
		return
	}

	ctx.String(200, response)
}
