// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

var ignoreLoggingPaths = map[string]bool{}

func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := ctx.Request.Method
		reqUri := ctx.Request.RequestURI
		statusCode := ctx.Writer.Status()

		if len(ctx.Errors) > 0 {
			slog.Error("API ERROR",
				"method", reqMethod,
				"URI", reqUri,
				"status", statusCode,
				"latency", latencyTime,
				"error", ctx.Errors.String(),
			)
		} else {
			fullPath := ctx.FullPath()
			if ignoreLoggingPaths[fullPath] {
				slog.Debug("API REQUEST",
					"method", reqMethod,
					"URI", reqUri,
					"status", statusCode,
					"latency", latencyTime,
				)
			} else {
				slog.Info("API REQUEST",
					"method", reqMethod,
					"URI", reqUri,
					"status", statusCode,
					"latency", latencyTime,
				)
			}
		}
	}
}
