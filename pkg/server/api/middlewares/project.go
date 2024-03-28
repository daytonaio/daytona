// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/server/auth"
	"github.com/gin-gonic/gin"
)

func ProjectAuthMiddleware() gin.HandlerFunc {
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

		if !auth.IsProjectApiKey(token) {
			ctx.AbortWithError(401, errors.New("unauthorized"))
		}

		ctx.Next()
	}
}
