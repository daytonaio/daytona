// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"context"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TelemetryMiddleware(telemetryService telemetry.TelemetryService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if telemetryService == nil {
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

		ctx.Next()
	}
}
