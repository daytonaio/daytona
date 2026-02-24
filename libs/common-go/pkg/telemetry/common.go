// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import "go.opentelemetry.io/otel/sdk/trace"

type Config struct {
	Endpoint       string
	Headers        map[string]string
	ServiceName    string
	ServiceVersion string
	Environment    string
}

type ExporterFilter interface {
	Apply(exporter trace.SpanExporter) trace.SpanExporter
}
