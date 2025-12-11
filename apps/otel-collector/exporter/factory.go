// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package exporter

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/daytonaio/apiclient"
	common_cache "github.com/daytonaio/common-go/pkg/cache"
	"github.com/daytonaio/otel-collector/exporter/internal/config"
)

const (
	typeStr   = "daytona_exporter"
	stability = component.StabilityLevelBeta
)

// NewFactory creates a factory for the custom exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		exporter.WithTraces(createTracesExporter, stability),
		exporter.WithMetrics(createMetricsExporter, stability),
		exporter.WithLogs(createLogsExporter, stability),
	)
}

// createDefaultConfig creates the default configuration for the exporter.
func createDefaultConfig() component.Config {
	return &Config{
		SandboxAuthTokenHeader: "sandbox-auth-token",
		CacheTTL:               5 * time.Minute,
		DefaultTimeout:         30 * time.Second,
		RetrySettings:          configretry.NewDefaultBackOffConfig(),
	}
}

// createTracesExporter creates a new trace exporter.
func createTracesExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	c := cfg.(*Config)

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: c.ApiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+c.ApiKey)
	apiClient := apiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	var cache common_cache.ICache[apiclient.OtelConfig]

	if c.Redis != nil {
		redisCache, err := common_cache.NewRedisCache[apiclient.OtelConfig](c.Redis, "org-otel-config:")
		if err != nil {
			return nil, err
		}
		cache = redisCache
	} else {
		cache = common_cache.NewMapCache[apiclient.OtelConfig]()
	}

	resolver := config.NewResolver(cache, set.Logger, apiClient)

	te := newTracesExporter(exporterConfig{
		config:   c,
		resolver: resolver,
		logger:   set.Logger,
	})

	return exporterhelper.NewTraces(
		ctx,
		set,
		cfg,
		te.push,
		exporterhelper.WithRetry(c.RetrySettings),
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: c.DefaultTimeout}),
		exporterhelper.WithShutdown(te.shutdown),
	)
}

// createMetricsExporter creates a new metrics exporter.
func createMetricsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	c := cfg.(*Config)

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: c.ApiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+c.ApiKey)
	apiClient := apiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	// Create cache and resolver
	memCache := common_cache.NewMapCache[apiclient.OtelConfig]()
	resolver := config.NewResolver(memCache, set.Logger, apiClient)

	me := newMetricExporter(exporterConfig{
		config:   c,
		resolver: resolver,
		logger:   set.Logger,
	})

	return exporterhelper.NewMetrics(
		ctx,
		set,
		cfg,
		me.push,
		exporterhelper.WithRetry(c.RetrySettings),
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: c.DefaultTimeout}),
		exporterhelper.WithShutdown(me.shutdown),
	)
}

// createLogsExporter creates a new logs exporter.
func createLogsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	c := cfg.(*Config)

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: c.ApiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+c.ApiKey)
	apiClient := apiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	// Create cache and resolver
	memCache := common_cache.NewMapCache[apiclient.OtelConfig]()
	resolver := config.NewResolver(memCache, set.Logger, apiClient)

	le := newLogsExporter(exporterConfig{
		config:   c,
		resolver: resolver,
		logger:   set.Logger,
	})

	return exporterhelper.NewLogs(
		ctx,
		set,
		cfg,
		le.push,
		exporterhelper.WithRetry(c.RetrySettings),
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: c.DefaultTimeout}),
		exporterhelper.WithShutdown(le.shutdown),
	)
}
