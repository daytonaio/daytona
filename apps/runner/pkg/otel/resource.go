// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package otel

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// newResource builds a deliberately minimal OTel resource. Every attribute
// here becomes a label on every exported series/span/log record, so we only
// include what's actually useful for grouping in dashboards and queries:
//
//   - service.name / service.namespace / service.version → routing + version
//   - deployment.environment.name                        → stage vs prod
//   - host.name / host.id                                → identify the runner
//
// The OS/SDK/process/runtime detectors are intentionally omitted to keep the
// per-series label set small. Operators can still inject extra attributes via
// the standard OTEL_RESOURCE_ATTRIBUTES env var (e.g. region, cluster).
func newResource(ctx context.Context, serviceName, serviceVersion, environment string) (*resource.Resource, error) {
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithHostID(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.ServiceNamespace("runner"),
			semconv.DeploymentEnvironmentName(environment),
		),
	)
	if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
		return res, nil
	}
	if err != nil {
		return nil, fmt.Errorf("otel: build resource: %w", err)
	}
	return res, nil
}
