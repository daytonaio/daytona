// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
	"unicode"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// otelState holds all OpenTelemetry infrastructure for the SDK.
type otelState struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	tracer         trace.Tracer
	meter          metric.Meter

	mu         sync.Mutex
	histograms map[string]metric.Float64Histogram
}

// getHistogram returns a cached histogram for the given metric name, creating one if needed.
func (s *otelState) getHistogram(name string) (metric.Float64Histogram, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if h, ok := s.histograms[name]; ok {
		return h, nil
	}

	h, err := s.meter.Float64Histogram(name,
		metric.WithUnit("ms"),
		metric.WithDescription(fmt.Sprintf("Duration of %s in milliseconds", name)),
	)
	if err != nil {
		return nil, err
	}
	s.histograms[name] = h
	return h, nil
}

// initOtel creates and configures the OpenTelemetry providers.
func initOtel(ctx context.Context) (*otelState, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("daytona-go-sdk"),
			semconv.ServiceVersion(Version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel resource: %w", err)
	}

	// Trace exporter (HTTP, gzip)
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	// Metric exporter
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
	)
	if err != nil {
		// Clean up trace provider on failure
		_ = tp.Shutdown(ctx)
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)

	// Set global propagator for W3C TraceContext
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("daytona-go-sdk")
	meter := mp.Meter("daytona-go-sdk")

	return &otelState{
		tracerProvider: tp,
		meterProvider:  mp,
		tracer:         tracer,
		meter:          meter,
		histograms:     make(map[string]metric.Float64Histogram),
	}, nil
}

// shutdownOtel flushes and shuts down both providers.
func shutdownOtel(ctx context.Context, state *otelState) error {
	if state == nil {
		return nil
	}

	var firstErr error
	if err := state.tracerProvider.Shutdown(ctx); err != nil {
		firstErr = err
	}
	if err := state.meterProvider.Shutdown(ctx); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

// otelTransport wraps an http.RoundTripper to inject trace context headers.
type otelTransport struct {
	base http.RoundTripper
}

func (t *otelTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Inject traceparent / tracestate headers
	otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
	return t.base.RoundTrip(req)
}

// withInstrumentation wraps a function with OpenTelemetry span creation and duration metrics.
// When state is nil (OTel disabled), fn is called directly with zero overhead.
func withInstrumentation[T any](ctx context.Context, state *otelState, component, method string, fn func(ctx context.Context) (T, error)) (T, error) {
	if state == nil {
		return fn(ctx)
	}

	spanName := component + "." + method
	ctx, span := state.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("component", component),
			attribute.String("method", method),
		),
	)
	defer span.End()

	start := time.Now()
	result, err := fn(ctx)
	duration := float64(time.Since(start).Milliseconds())

	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
	}

	metricName := toSnakeCase(spanName) + "_duration"
	if h, hErr := state.getHistogram(metricName); hErr == nil {
		h.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("component", component),
				attribute.String("method", method),
				attribute.String("status", status),
			),
		)
	}

	return result, err
}

// withInstrumentationVoid wraps an error-only function with OpenTelemetry instrumentation.
func withInstrumentationVoid(ctx context.Context, state *otelState, component, method string, fn func(ctx context.Context) error) error {
	_, err := withInstrumentation(ctx, state, component, method, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}

// toSnakeCase converts a PascalCase or camelCase string to snake_case.
// Dots are replaced with underscores for Prometheus-friendly metric names.
func toSnakeCase(s string) string {
	result := make([]byte, 0, len(s)*2)
	for i, r := range s {
		if r == '.' {
			result = append(result, '_')
			continue
		}
		if unicode.IsUpper(r) && i > 0 {
			prev := rune(s[i-1])
			if prev != '.' && !unicode.IsUpper(prev) {
				result = append(result, '_')
			}
		}
		result = append(result, byte(unicode.ToLower(r)))
	}
	return string(result)
}
