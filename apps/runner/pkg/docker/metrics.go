// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"strings"
	"time"

	otelpkg "github.com/daytonaio/runner/pkg/otel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Tracer for docker package spans. Reuses the shared name constant so traces
// land under the same instrumentation scope across the codebase.
var tracer = otel.Tracer(otelpkg.TracerDocker)

// Meter and instruments for docker registry operations (pulls/pushes against
// the snapshot registry). These are pushed via OTLP to the central collector
// alongside the existing Prometheus collectors in pkg/common/metrics.go.
var meter = otel.Meter(otelpkg.TracerDocker)

const (
	// RegistryOpPull labels an image pull from a registry.
	RegistryOpPull = "pull"
	// RegistryOpPush labels an image push to a registry.
	RegistryOpPush = "push"

	registryOpStatusSuccess = "success"
	registryOpStatusFailure = "failure"
)

var (
	// registryOpDuration records the wall-clock duration of registry operations.
	registryOpDuration metric.Float64Histogram
	// registryOpCount counts registry operations, partitioned by status.
	registryOpCount metric.Int64Counter
	// registryOpErrors counts only failed registry operations (convenience
	// signal for alerting; equivalent to registryOpCount{status="failure"}).
	registryOpErrors metric.Int64Counter
)

func init() {
	var err error

	registryOpDuration, err = meter.Float64Histogram("docker.registry.operation.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Wall-clock duration of Docker registry pull/push operations"))
	if err != nil {
		panic(err)
	}

	registryOpCount, err = meter.Int64Counter("docker.registry.operation.count",
		metric.WithUnit("{operation}"),
		metric.WithDescription("Count of Docker registry pull/push operations partitioned by status"))
	if err != nil {
		panic(err)
	}

	registryOpErrors, err = meter.Int64Counter("docker.registry.operation.errors",
		metric.WithUnit("{error}"),
		metric.WithDescription("Count of failed Docker registry pull/push operations"))
	if err != nil {
		panic(err)
	}
}

// RecordRegistryOp emits the count, duration, and (on error) error counter for
// a single Docker registry pull or push.
//
// sandboxID is attached as the sandbox.id attribute only when non-nil and
// non-empty, so high-cardinality blowup is avoided for snapshot/build flows
// that have no sandbox context.
//
// image.ref is attached as a label so operators can slice by image. This is
// the highest-cardinality label on these instruments (≈ one series per unique
// snapshot tag). If active-series cost becomes a problem, the lever is to
// drop the image.ref line from baseAttrs below and rely on the span (which
// always carries image.ref) for per-image attribution.
func RecordRegistryOp(ctx context.Context, operation string, sandboxID *string, imageRef string, start time.Time, err error) {
	status := registryOpStatusSuccess
	if err != nil {
		status = registryOpStatusFailure
	}

	registryHost := extractRegistryHost(imageRef)

	baseAttrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("registry.host", registryHost),
		attribute.String("image.ref", imageRef),
	}
	if sandboxID != nil && *sandboxID != "" {
		baseAttrs = append(baseAttrs, attribute.String("sandbox.id", *sandboxID))
	}

	durationAttrs := metric.WithAttributes(baseAttrs...)
	countAttrs := metric.WithAttributes(append(baseAttrs, attribute.String("status", status))...)

	registryOpDuration.Record(ctx, time.Since(start).Seconds(), durationAttrs)
	registryOpCount.Add(ctx, 1, countAttrs)
	if err != nil {
		registryOpErrors.Add(ctx, 1, countAttrs)
	}
}

// StartRegistrySpan opens a span for a Docker registry operation and seeds it
// with the common attributes used across pull/push call sites. Callers should
// defer span.End(); on error they should also call span.RecordError(err) before
// returning.
func StartRegistrySpan(ctx context.Context, name, operation string, sandboxID *string, imageRef string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("image.ref", imageRef),
		attribute.String("registry.host", extractRegistryHost(imageRef)),
	}
	if sandboxID != nil && *sandboxID != "" {
		attrs = append(attrs, attribute.String("sandbox.id", *sandboxID))
	}
	return tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// extractRegistryHost returns the registry hostname portion of an image
// reference, or "docker.io" when the reference is unqualified. The result is
// only used as a low-cardinality metric label (set of distinct registries the
// runner talks to), so heuristic parsing is acceptable here.
func extractRegistryHost(imageRef string) string {
	ref := strings.TrimPrefix(strings.TrimPrefix(imageRef, "https://"), "http://")

	slash := strings.IndexByte(ref, '/')
	if slash == -1 {
		return "docker.io"
	}

	first := ref[:slash]
	if strings.ContainsAny(first, ".:") || first == "localhost" {
		return first
	}
	return "docker.io"
}
