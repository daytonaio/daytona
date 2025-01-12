// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/api/controllers/apikey/dto"
	"github.com/daytonaio/daytona/pkg/api/util"
	"github.com/daytonaio/daytona/pkg/server"
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

	bearerToken := ctx.GetHeader("Authorization")
	if bearerToken == "" {
		ctx.AbortWithError(http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	token := util.ExtractToken(bearerToken)
	if token == "" {
		ctx.AbortWithError(http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	server := server.GetInstance(nil)

	currentApiKeyName, err := server.ApiKeyService.GetApiKeyName(ctx.Request.Context(), token)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("failed to get current api key name: %w", err))
		return
	}

	response, err := server.ApiKeyService.ListClientKeys(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get client API keys: %w", err))
		return
	}

	var result []dto.ApiKeyViewDTO
	for _, key := range response {
		viewKey := dto.ApiKeyViewDTO{
			Type: key.Type,
			Name: key.Name,
		}

		if viewKey.Name == currentApiKeyName {
			viewKey.Current = true
		}

		result = append(result, viewKey)
	}

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

	bearerToken := ctx.GetHeader("Authorization")
	if bearerToken == "" {
		ctx.AbortWithError(http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	token := util.ExtractToken(bearerToken)
	if token == "" {
		ctx.AbortWithError(http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	server := server.GetInstance(nil)

	currentApiKeyName, err := server.ApiKeyService.GetApiKeyName(ctx.Request.Context(), token)
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
