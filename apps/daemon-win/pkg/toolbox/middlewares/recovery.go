// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/daytonaio/daemon-win/pkg/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Recovery returns a middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"stack": string(debug.Stack()),
				}).Error("Panic recovered")

				errorResponse := common.ErrorResponse{
					StatusCode: http.StatusInternalServerError,
					Message:    "Internal server error",
					Code:       http.StatusText(http.StatusInternalServerError),
					Timestamp:  time.Now(),
					Path:       ctx.Request.URL.Path,
					Method:     ctx.Request.Method,
				}

				ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
			}
		}()
		ctx.Next()
	}
}
