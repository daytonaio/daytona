// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"context"
	"strings"
	"time"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var ignoreTelemetryPaths = map[string]bool{
	"/health":                             true,
	"/target/:targetId/metadata":          true,
	"/target/:targetId":                   true,
	"/workspace/:workspaceId/metadata":    true,
	"/workspace/:workspaceId":             true,
	"/runner/:runnerId/metadata":          true,
	"/runner/:runnerId":                   true,
	"/server/network-key":                 true,
	"/job/":                               true,
	"/runner/:runnerId/jobs":              true,
	"/runner/:runnerId/jobs/:jobId/state": true,
}

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

		reqUri := ctx.FullPath()
		if ignoreTelemetryPaths[reqUri] {
			ctx.Next()
			return
		}

		clientId := ctx.GetHeader(telemetry.CLIENT_ID_HEADER)
		if clientId == "" {
			clientId = uuid.NewString()
		}

		sessionId := ctx.GetHeader(telemetry.SESSION_ID_HEADER)
		if sessionId == "" {
			sessionId = internal.SESSION_ID
		}

		server := server.GetInstance(nil)

		telemetryCtx := context.WithValue(ctx.Request.Context(), telemetry.ENABLED_CONTEXT_KEY, true)
		telemetryCtx = context.WithValue(telemetryCtx, telemetry.CLIENT_ID_CONTEXT_KEY, clientId)
		telemetryCtx = context.WithValue(telemetryCtx, telemetry.SESSION_ID_CONTEXT_KEY, sessionId)
		telemetryCtx = context.WithValue(telemetryCtx, telemetry.SERVER_ID_CONTEXT_KEY, server.Id)

		ctx.Request = ctx.Request.WithContext(telemetryCtx)

		source := ctx.GetHeader(telemetry.SOURCE_HEADER)

		reqMethod := ctx.Request.Method

		query := ctx.Request.URL.RawQuery

		remoteProfile := false
		if source == string(telemetry.CLI_SOURCE) && !strings.Contains(ctx.Request.Host, "localhost") {
			remoteProfile = true
		}

		err := telemetryService.TrackServerEvent(telemetry.ServerEventApiRequestStarted, clientId, map[string]interface{}{
			"method":         reqMethod,
			"URI":            reqUri,
			"query":          query,
			"source":         source,
			"server_id":      server.Id,
			"session_id":     sessionId,
			"remote_profile": remoteProfile,
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
			"remote_profile": remoteProfile,
		}

		if len(ctx.Errors) > 0 {
			properties["error"] = ctx.Errors.String()
		}

		err = telemetryService.TrackServerEvent(telemetry.ServerEventApiResponseSent, clientId, properties)
		if err != nil {
			log.Trace(err)
		}

		ctx.Next()
	}
}
