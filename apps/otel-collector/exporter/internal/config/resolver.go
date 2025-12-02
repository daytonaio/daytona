package config

import (
	"context"
	"time"

	"github.com/daytonaio/apiclient"
	common_cache "github.com/daytonaio/common-go/pkg/cache"
	"github.com/daytonaio/otel-collector/exporter/internal/cache"
	"go.uber.org/zap"
)

// Resolver handles retrieving and caching endpoint configurations.
type Resolver struct {
	cache              cache.Cache
	logger             *zap.Logger
	authTokenEndpointCache common_cache.ICache[apiclient.OtelConfig]
	apiclient              *apiclient.APIClient
}

// NewResolver creates a new configuration resolver.
func NewResolver(cache cache.Cache, logger *zap.Logger, apiClient *apiclient.APIClient) *Resolver {
	return &Resolver{
		cache:  cache,
		logger: logger,
		apiclient: apiClient,
	}
}

func (r *Resolver) GetOrganizationOtelConfig(ctx context.Context, authToken string) (*apiclient.OtelConfig, error) {
	has, err := r.authTokenEndpointCache.Has(ctx, authToken)
	if err != nil {
		return nil, err
	}

	if has {
		return r.authTokenEndpointCache.Get(ctx, authToken)
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

	if err := r.authTokenEndpointCache.Set(ctx, authToken, *config, 10*time.Minute); err != nil {
		return nil, err
	}

	return config, nil
}

// InvalidateCache removes a specific sandbox configuration from the cache.
func (r *Resolver) InvalidateCache(ctx context.Context, sandboxID string) error {
	return r.cache.Delete(ctx, sandboxID)
}

// ClearCache removes all cached configurations.
func (r *Resolver) ClearCache(ctx context.Context) error {
	return r.cache.Clear(ctx)
}
