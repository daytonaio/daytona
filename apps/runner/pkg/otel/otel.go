// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package otel bootstraps OpenTelemetry tracing, metrics, and logging for the
// runner using the standard OTel environment variables. The shape mirrors the
// runner-vm pkg/otel package so both services export to the same collector
// with the same conventions.
package otel

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Option configures Init.
type Option func(*options)

type options struct {
	spanExporterWrappers []func(sdktrace.SpanExporter) sdktrace.SpanExporter
}

// WithSpanExporterWrapper appends a wrapper around the OTLP span exporter.
// Wrappers are applied in registration order; the outermost wrapper is the
// last one added. Useful for filters that mutate exported spans (e.g. the
// 404 status downgrade filter in pkg/telemetry/filters).
func WithSpanExporterWrapper(w func(sdktrace.SpanExporter) sdktrace.SpanExporter) Option {
	return func(o *options) { o.spanExporterWrappers = append(o.spanExporterWrappers, w) }
}

// ParseLogLevel parses a textual log level into slog.Level, defaulting to Info.
func ParseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Init initializes OpenTelemetry tracing, metrics, and logging.
// All configuration is via standard OTel env vars:
//   - OTEL_EXPORTER_OTLP_ENDPOINT — collector address (required to enable)
//   - OTEL_EXPORTER_OTLP_PROTOCOL — grpc (default) or http/protobuf
//   - OTEL_SDK_DISABLED=true — disable all signals
//   - OTEL_TRACES_EXPORTER=none — disable traces
//   - OTEL_METRICS_EXPORTER=none — disable metrics
//   - OTEL_LOGS_EXPORTER=none — disable logs
//
// When OTel is disabled, console-only slog logging is still configured.
// The returned shutdown closure flushes and shuts down every initialized
// signal provider; it is safe to call once at process exit.
func Init(ctx context.Context, serviceName, serviceVersion, environment string, logLevel slog.Level, opts ...Option) (func(context.Context) error, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	initConsoleLogging(logLevel)

	if strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true") ||
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		slog.Debug("otel disabled (no endpoint configured)")
		return noop, nil
	}

	slog.Info("otel enabled", "endpoint", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))

	res, err := newResource(ctx, serviceName, serviceVersion, environment)
	if err != nil {
		return nil, err
	}

	var shutdowns []func(context.Context) error

	if !signalDisabled("OTEL_TRACES_EXPORTER") {
		s, err := initTracing(ctx, res, o.spanExporterWrappers)
		if err != nil {
			return nil, err
		}
		shutdowns = append(shutdowns, s)
	}

	if !signalDisabled("OTEL_LOGS_EXPORTER") {
		s, err := initOTelLogging(ctx, res, serviceName, logLevel)
		if err != nil {
			return nil, err
		}
		shutdowns = append(shutdowns, s)
	}

	if !signalDisabled("OTEL_METRICS_EXPORTER") {
		s, err := initMetrics(ctx, res)
		if err != nil {
			return nil, err
		}
		shutdowns = append(shutdowns, s)
	}

	shutdown := func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdowns {
			errs = append(errs, fn(ctx))
		}
		return errors.Join(errs...)
	}
	return shutdown, nil
}

func signalDisabled(envVar string) bool {
	return strings.EqualFold(os.Getenv(envVar), "none")
}

func noop(context.Context) error { return nil }
