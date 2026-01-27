// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type OtelTracingConfig struct {
	OtelTracingEnabled  bool
	OtelSampleRate      float64
	OtelBatchTimeout    time.Duration
	OtelMaxBatchSize    int
	OtlpExporterTimeout time.Duration
	Environment         string
}

// InitTracing initializes OpenTelemetry tracing
func InitTracing(cfg OtelTracingConfig) (func(), error) {
	if !cfg.OtelTracingEnabled {
		// Return a no-op shutdown function when tracing is disabled
		return func() {}, nil
	}

	// Configure OTEL error handler to use slog at ERROR level
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.Error("OpenTelemetry error", "error", err)
	}))

	// Create resource with service information
	res, err := getOtelResource(cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP exporter
	ctx, cancel := context.WithTimeout(context.Background(), cfg.OtlpExporterTimeout)
	defer cancel()
	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(cfg.OtelBatchTimeout),
			sdktrace.WithMaxExportBatchSize(cfg.OtelMaxBatchSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.OtelSampleRate))),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return shutdown function
	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down trace provider: %v\n", err)
		}
	}

	return shutdown, nil
}
