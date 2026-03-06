// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type Config struct {
	Endpoint       string
	Headers        map[string]string
	ServiceName    string
	ServiceVersion string
	Environment    string
	ExtraLabels    map[string]string
}

func (c Config) Attributes() []attribute.KeyValue {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}

	attributes := []attribute.KeyValue{
		semconv.ServiceName(c.ServiceName),
		semconv.ServiceVersion(c.ServiceVersion),
		semconv.ServiceInstanceID(hostname),
		semconv.DeploymentEnvironmentName(c.Environment),
	}

	for k, v := range c.ExtraLabels {
		attributes = append(attributes, attribute.String(k, v))
	}

	return attributes
}

type ExporterFilter interface {
	Apply(exporter trace.SpanExporter) trace.SpanExporter
}
