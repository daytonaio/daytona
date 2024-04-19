// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func TelemetryMiddleware(telemetryService telemetry.TelemetryService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if telemetryService == nil {
			ctx.Next()
			return
		}

		if ctx.GetHeader(telemetry.ENABLED_HEADER) != "true" {
			ctx.Next()
			return
		}

		cliId := ctx.GetHeader(telemetry.CLI_ID_HEADER)
		if cliId == "" {
			cliId = uuid.NewString()
		}

		sessionId := ctx.GetHeader(telemetry.SESSION_ID_HEADER)
		if sessionId == "" {
			sessionId = internal.SESSION_ID
		}

		server := server.GetInstance(nil)

		telemetryCtx := context.WithValue(ctx.Request.Context(), telemetry.ENABLED_CONTEXT_KEY, true)
		telemetryCtx = context.WithValue(telemetryCtx, telemetry.CLI_ID_CONTEXT_KEY, cliId)
		telemetryCtx = context.WithValue(telemetryCtx, telemetry.SESSION_ID_CONTEXT_KEY, sessionId)
		telemetryCtx = context.WithValue(telemetryCtx, telemetry.SERVER_ID_CONTEXT_KEY, server.Id)

		ctx.Request = ctx.Request.WithContext(telemetryCtx)

		source := ctx.GetHeader(telemetry.SOURCE_HEADER)

		reqMethod := ctx.Request.Method
		reqUri := ctx.FullPath()

		query := ctx.Request.URL.RawQuery

		err := telemetryService.TrackServerEvent(telemetry.ServerEventApiRequestStarted, cliId, map[string]interface{}{
			"method":     reqMethod,
			"URI":        reqUri,
			"query":      query,
			"source":     source,
			"server_id":  server.Id,
			"session_id": sessionId,
		})
		if err != nil {
			log.Trace(err)
		}

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
			"source":         source,
			"exec time (µs)": execTime.Microseconds(),
			"server_id":      server.Id,
			"session_id":     sessionId,
		}

		if len(ctx.Errors) > 0 {
			properties["error"] = ctx.Errors.String()
		}

		err = telemetryService.TrackServerEvent(telemetry.ServerEventApiResponseSent, cliId, properties)
		if err != nil {
			log.Trace(err)
		}

		ctx.Next()
	}
}
