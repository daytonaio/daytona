// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// InitTracer initializes OpenTelemetry with Jaeger exporter
func InitTracer(ctx context.Context, config Config, exporterFilters ...ExporterFilter) (*trace.TracerProvider, error) {
	// Configure OTEL error handler to use slog
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.Error("OpenTelemetry error", "error", err)
	}))

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}

	// Create a new resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.ServiceInstanceID(hostname),
			semconv.DeploymentEnvironmentName(config.Environment),
		),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, err
	}

	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(config.Endpoint+"/v1/traces"),
		otlptracehttp.WithHeaders(config.Headers),
	)
	if err != nil {
		return nil, err
	}

	// This part allows users to provide custom filters for traceExporter
	// Order of filters is important - they will be applied in the order provided.
	// If no filter is provided, it will use the traceExporter as is.
	var spanExporter trace.SpanExporter = traceExporter
	for _, filter := range exporterFilters {
		spanExporter = filter.Apply(spanExporter)
	}

	// Create TracerProvider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(spanExporter),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tp, nil
}

// ShutdownTracer gracefully shuts down the tracer provider
func ShutdownTracer(logger *slog.Logger, tp *trace.TracerProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down tracer provider", "error", err)
	}
}
