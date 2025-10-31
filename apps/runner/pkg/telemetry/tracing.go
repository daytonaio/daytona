package telemetry

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/daytonaio/runner/cmd/runner/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracing initializes OpenTelemetry tracing
func InitTracing(cfg *config.Config) (func(), error) {
	if !cfg.OtelTracingEnabled {
		// Return a no-op shutdown function when tracing is disabled
		return func() {}, nil
	}

	// Configure OTEL error handler to use slog at ERROR level
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.Error("OpenTelemetry error", "error", err)
	}))

	// Create resource with service information
	res, err := getOtelResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient())
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(100),
		),
		sdktrace.WithResource(res),
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
