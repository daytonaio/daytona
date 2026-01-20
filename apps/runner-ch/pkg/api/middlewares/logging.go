// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()

		if path != "/" && path != "/metrics" {
			log.WithFields(log.Fields{
				"status":  statusCode,
				"method":  method,
				"path":    path,
				"latency": latency.String(),
				"ip":      c.ClientIP(),
			}).Info("Request")
		}
	}
}
