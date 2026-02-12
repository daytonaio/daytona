// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	Endpoint            string
	Headers             map[string]string
	ServiceName         string
	ServiceVersion      string
	Environment         string
	TraceExporterFilter func(*otlptrace.Exporter) sdktrace.SpanExporter
}
