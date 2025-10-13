// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package telemetry

import (
	"context"
	"time"

	"github.com/daytonaio/daemon/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	log "github.com/sirupsen/logrus"
)

// InitTracer initializes OpenTelemetry with Jaeger exporter
func InitTracer(ctx context.Context, config Config) (*sdktrace.TracerProvider, error) {
	// Create a new resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(internal.Version),
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
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// ShutdownTracer gracefully shuts down the tracer provider
func ShutdownTracer(tp *sdktrace.TracerProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	}
}
