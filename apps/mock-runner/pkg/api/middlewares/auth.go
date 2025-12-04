// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"net/http"
	"strings"

	"github.com/daytonaio/mock-runner/internal/constants"
	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(constants.AUTH_HEADER)
		if authHeader == "" {
			ctx.Error(common_errors.NewCustomError(http.StatusUnauthorized, "missing authorization header", "AUTH_HEADER_MISSING"))
			ctx.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, constants.AUTH_PREFIX) {
			ctx.Error(common_errors.NewCustomError(http.StatusUnauthorized, "invalid authorization format", "AUTH_FORMAT_INVALID"))
			ctx.Abort()
			return
		}

		/*
			TODO: implement later

			token := strings.TrimPrefix(authHeader, constants.AUTH_PREFIX)

			apiToken := os.Getenv("API_TOKEN")
			if apiToken == "" {
				ctx.Error(common_errors.NewCustomError(http.StatusInternalServerError, "API token not configured", "API_TOKEN_NOT_CONFIGURED"))
				ctx.Abort()
				return
			}

			if token != apiToken {
				ctx.Error(common_errors.NewCustomError(http.StatusUnauthorized, "invalid token " + token + " " + apiToken, "TOKEN_INVALID"))
				ctx.Abort()
				return
			}
		*/

		ctx.Next()
	}
}
