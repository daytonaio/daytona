// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/daytonaio/daytona/internal"
	"github.com/gin-gonic/gin"
)

const SERVER_VERSION_HEADER = "X-Server-Version"

func SetVersionMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer.Header().Add(SERVER_VERSION_HEADER, internal.Version)
	}
}
