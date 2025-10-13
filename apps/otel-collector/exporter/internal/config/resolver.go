// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"context"
	"time"

	common_cache "github.com/daytonaio/common-go/pkg/cache"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"go.uber.org/zap"
)

// Resolver handles retrieving and caching endpoint configurations.
type Resolver struct {
	cache     common_cache.ICache[apiclient.OtelConfig]
	logger    *zap.Logger
	apiclient *apiclient.APIClient
	cacheTTL  time.Duration
}

// NewResolver creates a new configuration resolver.
func NewResolver(cache common_cache.ICache[apiclient.OtelConfig], logger *zap.Logger, apiClient *apiclient.APIClient, cacheTTL time.Duration) *Resolver {
	return &Resolver{
		cache:     cache,
		logger:    logger,
		apiclient: apiClient,
		cacheTTL:  cacheTTL,
	}
}

func (r *Resolver) GetOrganizationOtelConfig(ctx context.Context, authToken string) (*apiclient.OtelConfig, error) {
	has, err := r.cache.Has(ctx, authToken)
	if err != nil {
		return nil, err
	}

	if has {
		otelConfig, err := r.cache.Get(ctx, authToken)
		if err != nil || otelConfig.Endpoint == "(none)" {
			return nil, err
		}
		return otelConfig, nil
	}

	otelConfig, res, err := r.apiclient.OrganizationsAPI.GetOrganizationOtelConfigBySandboxAuthToken(context.Background(), authToken).Execute()
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

	if err := r.cache.Set(ctx, authToken, *config, r.cacheTTL); err != nil {
		return nil, err
	}

	if config.Endpoint == "(none)" {
		return nil, nil
	}

	return config, nil
}

// InvalidateCache removes a specific sandbox configuration from the cache.
func (r *Resolver) InvalidateCache(ctx context.Context, sandboxID string) error {
	return r.cache.Delete(ctx, sandboxID)
}
