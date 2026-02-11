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

func (p *Proxy) Authenticate(ctx *gin.Context, sandboxIdOrSignedToken string, port float32) (sandboxId string, didRedirect bool, err error) {
	var authErrors []string

	// Try Authorization header with Bearer token
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		bearerToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		isValid, err := p.getSandboxBearerTokenValid(ctx, sandboxIdOrSignedToken, bearerToken)
		if err != nil {
			authErrors = append(authErrors, fmt.Sprintf("Bearer token validation error: %v", err))
		} else if isValid != nil && *isValid {
			return sandboxIdOrSignedToken, false, nil
		} else {
			authErrors = append(authErrors, "Bearer token is invalid")
		}
	}

	// Try auth key from header
	authKey := ctx.Request.Header.Get(SANDBOX_AUTH_KEY_HEADER)
	if authKey != "" {
		ctx.Request.Header.Del(SANDBOX_AUTH_KEY_HEADER)
		isValid, err := p.getSandboxAuthKeyValid(ctx, sandboxIdOrSignedToken, authKey)
		if err != nil {
			authErrors = append(authErrors, fmt.Sprintf("Auth key header validation error: %v", err))
		} else if isValid != nil && *isValid {
			return sandboxIdOrSignedToken, false, nil
		} else {
			authErrors = append(authErrors, "Auth key header is invalid")
		}
	}

	// Try auth key from query parameter
	queryAuthKey := ctx.Query(SANDBOX_AUTH_KEY_QUERY_PARAM)
	if queryAuthKey != "" {
		isValid, err := p.getSandboxAuthKeyValid(ctx, sandboxIdOrSignedToken, queryAuthKey)
		if err != nil {
			authErrors = append(authErrors, fmt.Sprintf("Auth key query param validation error: %v", err))
		} else if isValid != nil && *isValid {
			// Remove the auth key from the query string
			newQuery := ctx.Request.URL.Query()
			newQuery.Del(SANDBOX_AUTH_KEY_QUERY_PARAM)
			ctx.Request.URL.RawQuery = newQuery.Encode()
			return sandboxIdOrSignedToken, false, nil
		} else {
			authErrors = append(authErrors, "Auth key query parameter is invalid")
		}
	}

	// Try cookie authentication
	cookieSandboxId, err := ctx.Cookie(SANDBOX_AUTH_COOKIE_NAME + sandboxIdOrSignedToken)
	if err == nil && cookieSandboxId != "" {
		decodedValue := ""
		err = p.secureCookie.Decode(SANDBOX_AUTH_COOKIE_NAME+sandboxIdOrSignedToken, cookieSandboxId, &decodedValue)
		if err != nil {
			authErrors = append(authErrors, fmt.Sprintf("Cookie decoding error: %v", err))
		} else if decodedValue != "" {
			// For regular sandbox IDs: decodedValue is the sandbox ID itself
			// For signed tokens: decodedValue is the resolved sandbox ID
			return decodedValue, false, nil
		}
	}

	cookieDomain := p.getCookieDomain(ctx.Request.Host)

	sandboxId, err = p.getSandboxIdFromSignedPreviewUrlToken(ctx, sandboxIdOrSignedToken, port, cookieDomain)
	if err == nil {
		return sandboxId, false, nil
	} else {
		authErrors = append(authErrors, err.Error())
	}

	// All authentication methods failed, redirect to auth URL
	authUrl, err := p.getAuthUrl(ctx, sandboxIdOrSignedToken)
	if err != nil {
		return sandboxIdOrSignedToken, false, fmt.Errorf("failed to get auth URL: %w", err)
	}

	ctx.Redirect(http.StatusTemporaryRedirect, authUrl)

	// Return error with details about what failed
	var errorMsg string
	if len(authErrors) > 0 {
		errorMsg = fmt.Sprintf("authentication failed:\n%s", strings.Join(authErrors, "\n;\n"))
	} else {
		errorMsg = "missing authentication: provide a preview access token (via header, query parameter, or cookie) or use an API key or JWT"
	}

	return sandboxIdOrSignedToken, true, errors.New(errorMsg)
}

func (p *Proxy) getSandboxIdFromSignedPreviewUrlToken(ctx *gin.Context, sandboxIdOrSignedToken string, port float32, cookieDomain string) (string, error) {
	sandboxId, _, err := p.apiclient.PreviewAPI.GetSandboxIdFromSignedPreviewUrlToken(ctx.Request.Context(), sandboxIdOrSignedToken, port).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get sandbox ID: %w. Is the token expired?", err)
	}

	// Use the token (sandboxIdOrSignedToken) as the cookie name key so that
	// subsequent requests with the same signed URL find this cookie.
	// The cookie value stores the resolved sandbox ID.
	encoded, err := p.secureCookie.Encode(SANDBOX_AUTH_COOKIE_NAME+sandboxIdOrSignedToken, sandboxId)
	if err != nil {
		return "", fmt.Errorf("failed to encode cookie: %w", err)
	}

	ctx.SetCookie(SANDBOX_AUTH_COOKIE_NAME+sandboxIdOrSignedToken, encoded, 3600, "/", cookieDomain, p.config.EnableTLS, true)

	return sandboxId, nil
}
