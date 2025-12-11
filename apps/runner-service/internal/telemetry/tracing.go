/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package telemetry

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// InitTracer initializes OpenTelemetry with Jaeger exporter
func InitTracer() (*sdktrace.TracerProvider, error) {

	ctx := context.Background()

	// Create a new resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("runner-service"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create Jaeger exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("jaeger:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	// CRITICAL: Set global propagator for trace context propagation in HTTP headers
	// This enables otelhttp to automatically inject/extract trace context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // W3C Trace Context format
		propagation.Baggage{},      // W3C Baggage format
	))

	return tp, nil
}

// ShutdownTracer gracefully shuts down the tracer provider
func ShutdownTracer(tp *sdktrace.TracerProvider) {
	if tp == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	}
}
