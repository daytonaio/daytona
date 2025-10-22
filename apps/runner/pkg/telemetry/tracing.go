// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracing initializes OpenTelemetry tracing.
func InitTracing(cfg OtelConfig) (func(), error) {
	if !cfg.TracingEnabled {
		return func() {}, nil
	}

	// Configure OTEL error handler to use slog
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.Error("OpenTelemetry error", "error", err)
	}))

	// Create resource with service information
	res, err := getOtelResource(cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	baseExporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient())
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Wrap exporter to filter out 404 errors (expected in optimistic error handling)
	exporter := &filtered404Exporter{
		next: baseExporter,
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
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

// filtered404Exporter filters out HTTP 404 errors from being marked as errors in traces.
// This is useful for optimistic error handling patterns where 404 responses are expected
// (e.g., checking if a resource exists before creating it).
type filtered404Exporter struct {
	next sdktrace.SpanExporter
}

func (e *filtered404Exporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	// Filter out spans with 404 errors - they're expected in optimistic error handling
	filteredSpans := make([]sdktrace.ReadOnlySpan, 0, len(spans))
	for _, span := range spans {
		// Skip spans that have 404 errors - they're part of normal optimistic flow
		if span.Status().Code == codes.Error && e.is404Error(span) {
			// Don't export this span - it's an expected condition, not an error
			continue
		}
		filteredSpans = append(filteredSpans, span)
	}
	return e.next.ExportSpans(ctx, filteredSpans)
}

func (e *filtered404Exporter) is404Error(s sdktrace.ReadOnlySpan) bool {
	// Check for HTTP 404 status code
	for _, attr := range s.Attributes() {
		if attr.Key == "http.status_code" {
			statusCode := attr.Value.AsInterface()
			// Check if status code is 404 (not found)
			if statusCode == int64(404) || statusCode == 404 {
				return true
			}
		}
	}
	return false
}

func (e *filtered404Exporter) Shutdown(ctx context.Context) error {
	return e.next.Shutdown(ctx)
}
