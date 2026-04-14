// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	otelConfig, err := r.cache.Get(ctx, authToken)
	if err == nil {
		if otelConfig.Endpoint == "(none)" {
			return nil, nil
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

func (r *Resolver) GetOrganizationOtelConfigByOrgId(ctx context.Context, orgId string) (*apiclient.OtelConfig, error) {
	cacheKey := "org:" + orgId
	otelConfig, err := r.cache.Get(ctx, cacheKey)
	if err == nil {
		if otelConfig.Endpoint == "(none)" {
			return nil, nil
		}
		return otelConfig, nil
	}

	cfg := r.apiclient.GetConfig()
	baseURL := cfg.Servers[0].URL
	url := fmt.Sprintf("%s/organizations/%s/otel-config", baseURL, orgId)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range cfg.DefaultHeader {
		req.Header.Set(key, value)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	config := &apiclient.OtelConfig{
		Endpoint: "(none)",
	}

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var apiConfig apiclient.OtelConfig
		if err := json.Unmarshal(body, &apiConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		config = &apiConfig
	} else if resp.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("unexpected status code %d from API", resp.StatusCode)
	}

	if err := r.cache.Set(ctx, cacheKey, *config, r.cacheTTL); err != nil {
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
