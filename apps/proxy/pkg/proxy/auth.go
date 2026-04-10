// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/gin-gonic/gin"
)

func (p *Proxy) Authenticate(ctx *gin.Context, sandboxIdOrSignedToken string, port float32) (sandboxId string, didRedirect bool, err error) {
	var authErrors []string

	// Try Authorization header with Bearer token
	bearerToken := p.getBearerToken(ctx)
	if bearerToken != "" {
		isValid, err := p.getSandboxBearerTokenValid(ctx, sandboxIdOrSignedToken, bearerToken)
		if err != nil {
			authErrors = append(authErrors, fmt.Sprintf("Bearer token validation error: %v", err))
		} else if isValid != nil && *isValid {
			// If authentication successful, remove the Authorization header to prevent it from being forwarded to the sandbox
			ctx.Request.Header.Del("Authorization")
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
		} else {
			return decodedValue, false, nil
		}
	}

	if !ctx.GetBool(IS_TOOLBOX_REQUEST_KEY) {
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
	}

	// Return error with details about what failed
	var errorMsg string
	if len(authErrors) > 0 {
		errorMsg = fmt.Sprintf("authentication failed: %s", strings.Join(authErrors, ","))
	} else {
		errorMsg = "missing authentication: provide a preview access token (via header, query parameter, or cookie) or use an API key or JWT"
	}

	return sandboxIdOrSignedToken, !ctx.GetBool(IS_TOOLBOX_REQUEST_KEY), common_errors.NewUnauthorizedError(errors.New(errorMsg))
}

func (p *Proxy) getBearerToken(ctx *gin.Context) string {
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	return ""
}

func (p *Proxy) getSandboxIdFromSignedPreviewUrlToken(ctx *gin.Context, sandboxIdOrSignedToken string, port float32, cookieDomain string) (string, error) {
	var sandboxId string
	err := utils.RetryWithExponentialBackoff(ctx.Request.Context(), "getSandboxIdFromSignedPreviewUrlToken", proxyMaxRetries, proxyBaseDelay, proxyMaxDelay, func() error {
		s, _, e := p.apiclient.PreviewAPI.GetSandboxIdFromSignedPreviewUrlToken(ctx.Request.Context(), sandboxIdOrSignedToken, port).Execute()
		sandboxId = s
		openapiErr := common_errors.ConvertOpenAPIError(e)

		if openapiErr != nil && !common_errors.IsRetryableOpenAPIError(openapiErr) {
			return &utils.NonRetryableError{Err: openapiErr}
		}

		return openapiErr
	})
	if err != nil {
		return "", err
	}

	encoded, err := p.secureCookie.Encode(SANDBOX_AUTH_COOKIE_NAME+sandboxIdOrSignedToken, sandboxId)
	if err != nil {
		return "", fmt.Errorf("failed to encode cookie: %w", err)
	}

	ctx.SetCookie(SANDBOX_AUTH_COOKIE_NAME+sandboxIdOrSignedToken, encoded, 3600, "/", cookieDomain, p.config.EnableTLS, true)

	return sandboxId, nil
}
