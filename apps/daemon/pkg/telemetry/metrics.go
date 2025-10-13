// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package telemetry

import (
	"context"
	"time"

	"github.com/daytonaio/daemon/internal"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	log "github.com/sirupsen/logrus"
)

// InitMetrics initializes OpenTelemetry Metrics with an OTLP HTTP exporter.
func InitMetrics(ctx context.Context, config Config) (*metric.MeterProvider, error) {
	// Resource describing this service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(internal.Version),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create OTLP HTTP metrics exporter
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(config.Endpoint+"/v1/metrics"),
		otlpmetrichttp.WithHeaders(config.Headers),
	)
	if err != nil {
		return nil, err
	}

	// Periodic reader to push metrics on an interval
	reader := metric.NewPeriodicReader(exporter)

	// MeterProvider with resource and reader
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader),
	)

	// Set as global provider so otel.Meter(...) uses it
	otel.SetMeterProvider(mp)

	// Start system metrics collection
	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
		log.Printf("Failed to start runtime metrics: %v", err)
	}

	// Start host metrics collection
	log.Info("Starting host metrics collection")
	if err := host.Start(host.WithMeterProvider(mp)); err != nil {
		log.Printf("Failed to start host metrics: %v", err)
	}

	return mp, nil
}

// ShutdownMeter gracefully shuts down the MeterProvider and flushes metrics.
func ShutdownMeter(mp *metric.MeterProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mp.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down meter provider: %v", err)
	}
}
