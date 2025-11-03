// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package telemetry

import (
	"context"
	"os"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func getOtelResource(cfg *config.Config) (*resource.Resource, error) {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}

	return resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("daytona-runner"),
			semconv.ServiceVersion(internal.Version),
			semconv.ServiceInstanceID(hostname),
			semconv.DeploymentEnvironmentName(cfg.Environment),
		),
		resource.WithTelemetrySDK(),
	)
}
