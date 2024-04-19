// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"time"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/gin-gonic/gin"
)

func TelemetryMiddleware(telemetryService *telemetry.TelemetryService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if telemetryService == nil {
			ctx.Next()
			return
		}

		telemetryEnabled := ctx.GetHeader(telemetry.ENABLED_HEADER)
		if telemetryEnabled != "true" {
			ctx.Next()
			return
		}

		sessionId := ctx.GetHeader(telemetry.SESSION_ID_HEADER)
		if sessionId == "" {
			sessionId = internal.SESSION_ID
		}

		reqMethod := ctx.Request.Method
		// ReqUri should not include path params
		reqUri := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery

		(*telemetryService).TrackServerEvent(telemetry.ServerEventApiRequestStarted, sessionId, map[string]interface{}{
			"method": reqMethod,
			"URI":    reqUri,
			"query":  query,
		})

		startTime := time.Now()
		ctx.Next()
		endTime := time.Now()
		execTime := endTime.Sub(startTime)
		statusCode := ctx.Writer.Status()

		properties := map[string]interface{}{
			"method":         reqMethod,
			"URI":            reqUri,
			"query":          query,
			"status":         statusCode,
			"exec time (µs)": execTime.Microseconds(),
		}

		if len(ctx.Errors) > 0 {
			properties["error"] = ctx.Errors.String()
		}

		(*telemetryService).TrackServerEvent(telemetry.ServerEventApiResponseSent, sessionId, properties)

		ctx.Next()
	}
}
