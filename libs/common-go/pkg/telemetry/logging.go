// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/daytonaio/common-go/pkg/log"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// InitLogger initializes OpenTelemetry logging with OTLP exporter.
// It wraps the provided slog.Logger with OTEL support using a fanout handler.
// Returns the logger (either the original or a new one with OTEL support) and a shutdown function.
func InitLogger(ctx context.Context, config Config) (*otellog.LoggerProvider, error) {
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
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP log exporter
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(config.Endpoint+"/v1/logs"),
		otlploghttp.WithHeaders(config.Headers),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	// Create LoggerProvider
	lp := otellog.NewLoggerProvider(
		otellog.WithProcessor(otellog.NewBatchProcessor(exporter)),
		otellog.WithResource(res),
	)

	// Set global LoggerProvider
	global.SetLoggerProvider(lp)

	// Create OTEL slog handler
	otelHandler := otelslog.NewHandler(config.ServiceName, otelslog.WithLoggerProvider(lp))

	// Get the default slog logger
	defaultLogger := slog.Default()

	// Wrap OTEL handler with level filter to respect the logger's configured level
	filteredOtelHandler := log.NewLevelFilterHandler(otelHandler, defaultLogger.Handler())

	// Create fanout handler combining existing logger's handler and filtered OTEL handler
	fanoutHandler := log.NewMultiHandler(
		[]slog.Handler{
			defaultLogger.Handler(),
			filteredOtelHandler,
		}...,
	)

	// Create new logger instance with fanout handler
	newLogger := slog.New(fanoutHandler)

	// Set as default logger globally
	slog.SetDefault(newLogger)

	return lp, nil
}
