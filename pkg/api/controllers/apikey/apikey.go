// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"fmt"
	"net/http"

	internal_util "github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/api/controllers/apikey/dto"
	"github.com/daytonaio/daytona/pkg/api/util"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/gin-gonic/gin"
)

// ListClientApiKeys 			godoc
//
//	@Tags			apiKey
//	@Summary		List API keys
//	@Description	List API keys
//	@Produce		json
//	@Success		200	{array}	ApiKeyViewDTO
//	@Router			/apikey [get]
//
//	@id				ListClientApiKeys
func ListClientApiKeys(ctx *gin.Context) {
	server := server.GetInstance(nil)

	currentApiKeyName, err := server.ApiKeyService.GetApiKeyName(ctx.Request.Context(), util.ExtractToken(ctx))
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to get current api key name: %w", err))
		return
	}

	response, err := server.ApiKeyService.ListClientKeys(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get client API keys: %w", err))
		return
	}

	result := internal_util.ArrayMap(response, func(key *services.ApiKeyDTO) dto.ApiKeyViewDTO {
		return dto.ApiKeyViewDTO{Name: key.Name, Type: key.Type, Current: key.Name == currentApiKeyName}
	})

	ctx.JSON(200, result)
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

	currentApiKeyName, err := server.ApiKeyService.GetApiKeyName(ctx.Request.Context(), util.ExtractToken(ctx))
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to get current api key name: %w", err))
		return
	}

	if currentApiKeyName == apiKeyName {
		ctx.AbortWithError(http.StatusForbidden, fmt.Errorf("cannot revoke current api key"))
		return
	}

	err = server.ApiKeyService.Revoke(ctx.Request.Context(), apiKeyName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to revoke api key: %w", err))
		return
	}

	ctx.Status(200)
}
