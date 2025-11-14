// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	log "github.com/sirupsen/logrus"
)

// InitTracer initializes OpenTelemetry with Jaeger exporter
func InitTracer(ctx context.Context, config Config) (*trace.TracerProvider, error) {
	// Create a new resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(config.Endpoint+"/v1/traces"),
		otlptracehttp.WithHeaders(config.Headers),
	)
	if err != nil {
		return nil, err
	}

	// Create TracerProvider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
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
func ShutdownTracer(tp *trace.TracerProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	}
}
