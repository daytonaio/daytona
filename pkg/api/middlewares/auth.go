// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"errors"
	"strings"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bearerToken := ctx.GetHeader("Authorization")
		if bearerToken == "" {
			ctx.AbortWithError(401, errors.New("unauthorized"))
			return
		}

		token := ExtractToken(bearerToken)
		if token == "" {
			ctx.AbortWithError(401, errors.New("unauthorized"))
			return
		}

		server := server.GetInstance(nil)

		if !server.ApiKeyService.IsValidApiKey(token) {
			ctx.AbortWithError(401, errors.New("unauthorized"))
			return
		}

		apiKeyType := models.ApiKeyTypeClient

		if server.ApiKeyService.IsTargetApiKey(token) {
			apiKeyType = models.ApiKeyTypeTarget
		} else if server.ApiKeyService.IsWorkspaceApiKey(token) {
			apiKeyType = models.ApiKeyTypeWorkspace
		}

		ctx.Set("apiKeyType", apiKeyType)
		ctx.Next()
	}
}

func ExtractToken(bearerToken string) string {
	if !strings.HasPrefix(bearerToken, "Bearer ") {
		return ""
	}

	return strings.TrimPrefix(bearerToken, "Bearer ")
}
