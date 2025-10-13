// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"

	"go.opentelemetry.io/contrib/bridges/otellogrus"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	log "github.com/sirupsen/logrus"
)

// InitLogger initializes OpenTelemetry logging with OTLP exporter
func InitLogger(ctx context.Context, config Config) (*otellog.LoggerProvider, error) {
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

	// Create OTLP log exporter
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(config.Endpoint+"/v1/logs"),
		otlploghttp.WithHeaders(config.Headers),
	)
	if err != nil {
		return nil, err
	}

	// Create LoggerProvider
	lp := otellog.NewLoggerProvider(
		otellog.WithProcessor(otellog.NewBatchProcessor(exporter)),
		otellog.WithResource(res),
	)

	// Set global LoggerProvider
	global.SetLoggerProvider(lp)

	log.AddHook(otellogrus.NewHook(config.ServiceName, otellogrus.WithLoggerProvider(lp)))

	return lp, nil
}
