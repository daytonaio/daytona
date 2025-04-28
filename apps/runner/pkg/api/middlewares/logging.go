// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

var ignoreLoggingPaths = map[string]bool{}

func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)

		fields := log.Fields{
			"method":  ctx.Request.Method,
			"URI":     ctx.Request.RequestURI,
			"status":  ctx.Writer.Status(),
			"latency": latencyTime,
		}

		// Determine log level based on status code and errors
		if ignoreLoggingPaths[ctx.FullPath()] {
			log.WithFields(fields).Debug("API REQUEST")
		} else {
			log.WithFields(fields).Info("API REQUEST")
		}
	}
}
