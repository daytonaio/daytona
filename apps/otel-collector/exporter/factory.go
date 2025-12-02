package exporter

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/daytonaio/otel-collector/exporter/internal/cache"
	"github.com/daytonaio/otel-collector/exporter/internal/config"
)

const (
	typeStr   = "daytonaexporter"
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
		SandboxIDHeader: "sandboxId",
		CacheTTL:        5 * time.Minute,
		DefaultTimeout:  30 * time.Second,
		RetrySettings:   configretry.NewDefaultBackOffConfig(),
	}
}

// createTracesExporter creates a new trace exporter.
func createTracesExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	c := cfg.(*Config)

	// Create cache and resolver
	memCache := cache.NewMemoryCache()
	resolver := config.NewResolver(memCache, set.Logger)

	te := &tracesExporter{
		config:   c,
		resolver: resolver,
		logger:   set.Logger,
	}

	return exporterhelper.NewTraces(
		ctx,
		set,
		cfg,
		te.pushTraces,
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

	// Create cache and resolver
	memCache := cache.NewMemoryCache()
	resolver := config.NewResolver(memCache, set.Logger)

	me := &metricsExporter{
		config:   c,
		resolver: resolver,
		logger:   set.Logger,
	}

	return exporterhelper.NewMetrics(
		ctx,
		set,
		cfg,
		me.pushMetrics,
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

	// Create cache and resolver
	memCache := cache.NewMemoryCache()
	resolver := config.NewResolver(memCache, set.Logger)

	le := &logsExporter{
		config:   c,
		resolver: resolver,
		logger:   set.Logger,
	}

	return exporterhelper.NewLogs(
		ctx,
		set,
		cfg,
		le.pushLogs,
		exporterhelper.WithRetry(c.RetrySettings),
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: c.DefaultTimeout}),
		exporterhelper.WithShutdown(le.shutdown),
	)
}
