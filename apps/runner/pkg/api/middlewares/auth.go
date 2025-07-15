// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"errors"
	"os"
	"strings"

	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(constants.DAYTONA_AUTHORIZATION_HEADER)
		if authHeader == "" {
			authHeader = ctx.GetHeader(constants.AUTHORIZATION_HEADER)
		}

		ctx.Request.Header.Del(constants.DAYTONA_AUTHORIZATION_HEADER)

		if authHeader == "" {
			ctx.Error(common.NewUnauthorizedError(errors.New("authorization header required")))
			ctx.Abort()
			return
		}

		// Split "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != constants.BEARER_AUTH_HEADER {
			ctx.Error(common.NewUnauthorizedError(errors.New("invalid authorization header format")))
			ctx.Abort()
			return
		}

		token := parts[1]
		// Compare with API token from environment variable
		if token != os.Getenv("API_TOKEN") {
			ctx.Error(common.NewUnauthorizedError(errors.New("invalid token")))
			ctx.Abort()
			return
		}

		// Authentication successful, continue to the next handler
		ctx.Next()
	}
}
