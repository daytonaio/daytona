// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (p *Proxy) Authenticate(ctx *gin.Context, sandboxId string) (err error, didRedirect bool) {
	var authErrors []string

	// Try Authorization header with Bearer token
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		bearerToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		isValid, err := p.getSandboxBearerTokenValid(ctx, sandboxId, bearerToken)
		if err != nil {
			authErrors = append(authErrors, fmt.Sprintf("Bearer token validation error: %v", err))
		} else if isValid != nil && *isValid {
			return nil, false
		} else {
			authErrors = append(authErrors, "Bearer token is invalid")
		}
	}

	// Try auth key from header
	authKey := ctx.Request.Header.Get(SANDBOX_AUTH_KEY_HEADER)
	if authKey != "" {
		isValid, err := p.getSandboxAuthKeyValid(ctx, sandboxId, authKey)
		if err != nil {
			authErrors = append(authErrors, fmt.Sprintf("Auth key header validation error: %v", err))
		} else if isValid != nil && *isValid {
			return nil, false
		} else {
			authErrors = append(authErrors, "Auth key header is invalid")
		}
	}

	// Try auth key from query parameter
	queryAuthKey := ctx.Query(SANDBOX_AUTH_KEY_QUERY_PARAM)
	if queryAuthKey != "" {
		isValid, err := p.getSandboxAuthKeyValid(ctx, sandboxId, queryAuthKey)
		if err != nil {
			authErrors = append(authErrors, fmt.Sprintf("Auth key query param validation error: %v", err))
		} else if isValid != nil && *isValid {
			// Remove the auth key from the query string
			newQuery := ctx.Request.URL.Query()
			newQuery.Del(SANDBOX_AUTH_KEY_QUERY_PARAM)
			ctx.Request.URL.RawQuery = newQuery.Encode()
			return nil, false
		} else {
			authErrors = append(authErrors, "Auth key query parameter is invalid")
		}
	}

	// Try cookie authentication
	cookieSandboxId, err := ctx.Cookie(SANDBOX_AUTH_COOKIE_NAME + sandboxId)
	if err == nil && cookieSandboxId != "" {
		decodedValue := ""
		err = p.secureCookie.Decode(SANDBOX_AUTH_COOKIE_NAME+sandboxId, cookieSandboxId, &decodedValue)
		if err != nil {
			authErrors = append(authErrors, fmt.Sprintf("Cookie decoding error: %v", err))
		} else if decodedValue == sandboxId {
			return nil, false
		} else {
			authErrors = append(authErrors, "Cookie sandbox ID mismatch")
		}
	}

	// All authentication methods failed, redirect to auth URL
	authUrl, err := p.getAuthUrl(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to get auth URL: %w", err), false
	}

	ctx.Redirect(http.StatusTemporaryRedirect, authUrl)

	// Return error with details about what failed
	errorMsg := "authentication failed"
	if len(authErrors) > 0 {
		errorMsg = fmt.Sprintf("authentication failed:\n%s", strings.Join(authErrors, "\n;\n"))
	} else {
		errorMsg = "missing authentication: provide a preview access token (via header, query parameter, or cookie) or use an API key or JWT"
	}

	return errors.New(errorMsg), true
}
