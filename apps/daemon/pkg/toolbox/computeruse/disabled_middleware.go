// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
)

// computerUseDisabledMiddleware returns a middleware that handles requests when computer-use is disabled
func ComputerUseDisabledMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Error(common_errors.NewCustomError(
			http.StatusServiceUnavailable,
			"computer-use plugin is unavailable: required X11 runtime libraries are missing on this sandbox; rebuild the sandbox image with the computer-use dependencies installed",
			"",
		))
		c.Abort()
	}
}
