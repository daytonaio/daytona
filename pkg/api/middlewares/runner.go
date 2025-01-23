// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/api/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

func RunnerAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := util.ExtractToken(ctx)
		if token == "" {
			ctx.AbortWithError(401, errors.New("unauthorized"))
			return
		}

		server := server.GetInstance(nil)

		apiKeyType, err := server.ApiKeyService.GetApiKeyType(ctx.Request.Context(), token)
		if err != nil {
			ctx.AbortWithError(401, errors.New("unauthorized"))
			return
		}

		if apiKeyType != models.ApiKeyTypeRunner {
			ctx.AbortWithError(401, errors.New("unauthorized"))
			return
		}

		ctx.Next()
	}
}
