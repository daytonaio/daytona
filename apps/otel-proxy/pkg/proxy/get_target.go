// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/daytonaio/apiclient"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func (p *Proxy) GetProxyTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	// Extract sandbox token from header
	sandboxAuthToken := ctx.GetHeader("sandbox-auth-token")
	if sandboxAuthToken == "" {
		ctx.Error(common_errors.NewUnauthorizedError(errors.New("sandbox auth token is missing")))
		return nil, nil, errors.New("sandbox auth token is missing")
	}

	ctx.Request.Header.Del("sandbox-auth-token")

	otelConfig, err := p.getOrganizationOtelConfig(ctx.Request.Context(), sandboxAuthToken)

	if err != nil {
		// Don't fail the request. Treat like there is no org otel endpoint
		log.Warn("Failed to get organization OTEL endpoint: ", err)
		return nil, nil, nil
	}

	if otelConfig == nil || otelConfig.Endpoint == "(none)" {
		// No org otel endpoint configured
		return nil, nil, nil
	}

	// Create the complete target URL with path
	target, err := url.Parse(otelConfig.Endpoint)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err)))
		return nil, nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	// add path to target
	target.Path = ctx.Request.URL.Path
	target.RawQuery = ctx.Request.URL.RawQuery

	return target, otelConfig.Headers, nil
}

func (p *Proxy) getOrganizationOtelConfig(ctx context.Context, authToken string) (*apiclient.OtelConfig, error) {
	has, err := p.authTokenEndpointCache.Has(ctx, authToken)
	if err != nil {
		return nil, err
	}

	if has {
		return p.authTokenEndpointCache.Get(ctx, authToken)
	}

	otelConfig, res, err := p.apiclient.OrganizationsAPI.GetOrganizationOtelConfigBySandboxAuthToken(context.Background(), authToken).Execute()
	if err != nil && res != nil && res.StatusCode != 404 {
		return nil, err
	}

	// Store this in cache to prevent repeated api calls for orgs that don't have otel endpoints
	config := &apiclient.OtelConfig{
		Endpoint: "(none)",
	}

	if otelConfig != nil {
		config = &apiclient.OtelConfig{
			Endpoint: otelConfig.Endpoint,
			Headers:  otelConfig.Headers,
		}
	}

	if err := p.authTokenEndpointCache.Set(ctx, authToken, *config, 10*time.Minute); err != nil {
		return nil, err
	}

	return config, nil
}
