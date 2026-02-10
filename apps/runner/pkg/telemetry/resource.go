// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package telemetry

import (
	"context"
	"os"

	"github.com/daytonaio/runner/internal"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// OtelConfig contains configuration for OpenTelemetry tracing and logging.
type OtelConfig struct {
	TracingEnabled bool
	LoggingEnabled bool
	Environment    string
}

func getOtelResource(environment string) (*resource.Resource, error) {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}

	return resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("daytona-runner"),
			semconv.ServiceVersion(internal.Version),
			semconv.ServiceInstanceID(hostname),
			semconv.DeploymentEnvironmentName(environment),
		),
		resource.WithTelemetrySDK(),
	)
}
