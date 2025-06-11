// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *Proxy) Authenticate(ctx *gin.Context, sandboxId string) (err error, didRedirect bool) {
	authKey := ctx.Request.Header.Get(DAYTONA_SANDBOX_AUTH_KEY_HEADER)
	if authKey == "" {
		if ctx.Query(DAYTONA_SANDBOX_AUTH_KEY_QUERY_PARAM) != "" {
			authKey = ctx.Query(DAYTONA_SANDBOX_AUTH_KEY_QUERY_PARAM)
			newQuery := ctx.Request.URL.Query()
			newQuery.Del(DAYTONA_SANDBOX_AUTH_KEY_QUERY_PARAM)
			ctx.Request.URL.RawQuery = newQuery.Encode()
		} else {
			// Check for cookie
			cookieSandboxId, err := ctx.Cookie(DAYTONA_SANDBOX_AUTH_COOKIE_NAME + sandboxId)
			if err == nil && cookieSandboxId != "" {
				decodedValue := ""
				err = p.secureCookie.Decode(DAYTONA_SANDBOX_AUTH_COOKIE_NAME+sandboxId, cookieSandboxId, &decodedValue)
				if err != nil {
					return errors.New("sandbox not found"), false
				}

				if decodedValue != sandboxId {
					return errors.New("sandbox not found"), false
				} else {
					return nil, false
				}
			} else {
				authUrl, err := p.getAuthUrl(ctx, sandboxId)
				if err != nil {
					return fmt.Errorf("failed to get auth URL: %w", err), false
				}

				ctx.Redirect(http.StatusTemporaryRedirect, authUrl)

				return errors.New("auth key is required"), true
			}
		}
	}

	if authKey != "" {
		isValid, err := p.getSandboxAuthKeyValid(ctx, sandboxId, authKey)
		if err != nil {
			return fmt.Errorf("failed to get sandbox auth key valid status: %w", err), false
		}

		if !*isValid {
			return errors.New("invalid auth key"), false
		} else {
			return nil, false
		}
	}

	return errors.New("auth key is required. Authenticate via a Header, Query Param or Cookie"), false
}
