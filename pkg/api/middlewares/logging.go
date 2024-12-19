// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

var ignoreLoggingPaths = map[string]bool{
	"/job/":                               true,
	"/workspace/:workspaceId/metadata":    true,
	"/target/:targetId/metadata":          true,
	"/runner/:runnerId/jobs":              true,
	"/runner/:runnerId/metadata":          true,
	"/runner/:runnerId/jobs/:jobId/state": true,
}

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
			log.WithFields(log.Fields{
				"method":  reqMethod,
				"URI":     reqUri,
				"status":  statusCode,
				"latency": latencyTime,
				"error":   ctx.Errors.String(),
			}).Error("API ERROR")
			ctx.JSON(statusCode, gin.H{"error": ctx.Errors[0].Err.Error()})
		} else {
			fullPath := ctx.FullPath()
			if ignoreLoggingPaths[fullPath] {
				log.WithFields(log.Fields{
					"method":  reqMethod,
					"URI":     reqUri,
					"status":  statusCode,
					"latency": latencyTime,
				}).Debug("API REQUEST")
			} else {
				log.WithFields(log.Fields{
					"method":  reqMethod,
					"URI":     reqUri,
					"status":  statusCode,
					"latency": latencyTime,
				}).Info("API REQUEST")
			}
		}

		ctx.Next()
	}
}
