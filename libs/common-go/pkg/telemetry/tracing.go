// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// InitTracer initializes OpenTelemetry with Jaeger exporter
func InitTracer(ctx context.Context, config Config) (*trace.TracerProvider, error) {
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

	// This part allows users to provide a custom filter function to wrap the traceExporter
	// Can be used to filter out certain errors (like 404s) that are expected in optimistic error handling scenarios.
	// If no filter is provided, it will use the traceExporter as is.
	var spanExporter sdktrace.SpanExporter = traceExporter
	if config.TraceExporterFilter != nil {
		spanExporter = config.TraceExporterFilter(traceExporter)
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
