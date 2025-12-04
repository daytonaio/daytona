// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		path := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery

		ctx.Next()

		endTime := time.Now()
		latency := endTime.Sub(startTime)
		method := ctx.Request.Method
		statusCode := ctx.Writer.Status()
		clientIP := ctx.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		log.WithFields(log.Fields{
			"status":  statusCode,
			"latency": latency.String(),
			"ip":      clientIP,
			"method":  method,
			"path":    path,
		}).Info("Request")
	}
}



