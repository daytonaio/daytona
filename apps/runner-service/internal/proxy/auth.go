// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"net/http"
	"strings"
)

const (
	DaytonaAuthorizationHeader = "X-Daytona-Authorization"
	AuthorizationHeader        = "Authorization"
	BearerPrefix               = "Bearer"
)

func withBearerAuth(next http.HandlerFunc, expectedToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for X-Daytona-Authorization
		authHeader := r.Header.Get(DaytonaAuthorizationHeader)

		// Remove X-Daytona-Authorization header after reading
		r.Header.Del(DaytonaAuthorizationHeader)

		// Check if header is present
		if authHeader == "" {
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>" format
		token, ok := strings.CutPrefix(authHeader, BearerPrefix+" ")
		if !ok {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Compare with expected token
		if token != expectedToken {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Authentication successful, continue to the next handler
		next(w, r)
	}
}
