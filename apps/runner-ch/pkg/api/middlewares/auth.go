// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader        = "Authorization"
	DaytonaAuthorizationHeader = "X-Daytona-Authorization"
	BearerPrefix               = "Bearer"
)

// AuthMiddleware validates the Bearer token
// Checks both X-Daytona-Authorization (from proxy) and Authorization headers
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := os.Getenv("API_TOKEN")

		if token == "" {
			// No token configured, allow all requests
			c.Next()
			return
		}

		// Check X-Daytona-Authorization first (from proxy service)
		// then fall back to Authorization header
		authHeader := c.GetHeader(DaytonaAuthorizationHeader)
		if authHeader == "" {
			authHeader = c.GetHeader(AuthorizationHeader)
		}

		// Remove the Daytona header to prevent forwarding to VM
		c.Request.Header.Del(DaytonaAuthorizationHeader)

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		if parts[1] != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Next()
	}
}
