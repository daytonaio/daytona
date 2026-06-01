// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/daytonaio/runner/internal/otlptest"
	"github.com/daytonaio/runner/pkg/docker"
	otelpkg "github.com/daytonaio/runner/pkg/otel"
	"github.com/stretchr/testify/require"
)

// TestRegistryMetricsExport is an end-to-end test of the Docker registry
// metrics pipeline: instrument → SDK → autoexport OTLP/HTTP exporter →
// in-process mock collector. It verifies that RecordRegistryOp emits the
// expected metric points with the expected attributes and resource info.
//
// Because pkg/otel.Init flips global OTel providers, this test runs as a
// single top-level function. Individual scenarios are split into t.Run
// subtests so failures attribute to a specific case.
func TestRegistryMetricsExport(t *testing.T) {
	srv := otlptest.New(t)

	// Force autoexport to pick the HTTP/protobuf exporter (default is gRPC),
	// disable the signals we don't exercise to avoid spurious 404s against
	// the mock server, and shorten the export interval so we don't sit on
	// the SDK's 60s default.
	// Reset inherited knobs that would otherwise disable metrics or perturb the
	// asserted resource labels in a developer/CI shell that exports OTel env.
	t.Setenv("OTEL_SDK_DISABLED", "")
	t.Setenv("OTEL_METRICS_EXPORTER", "")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_LOGS_EXPORTER", "none")
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "100")
	// Defensive: keep autoexport from picking up workspace-set OTLP headers
	// that would never reach our mock.
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "")

	ctx := context.Background()
	shutdown, err := otelpkg.Init(ctx, "daytona-runner", "test", "test", slog.LevelError)
	require.NoError(t, err, "pkg/otel.Init failed")

	// Guard against double-shutdown (defer + explicit flush below).
	shutdownOnce := sync.OnceValue(func() error {
		return shutdown(context.Background())
	})
	t.Cleanup(func() {
		if err := shutdownOnce(); err != nil {
			t.Logf("otel shutdown error: %v", err)
		}
	})

	// ── emit ─────────────────────────────────────────────────────────────
	// Five distinct cases exercised in one batch so a single shutdown flushes
	// all of them in one go. Start times are staggered by 10ms so the
	// histogram observations land in deterministic-ish buckets.
	const (
		pullSandboxOK     = "sb-pull-success"
		pushSandboxOK     = "sb-push-success"
		pushSandboxFailed = "sb-push-failure"
	)

	now := time.Now()
	docker.RecordRegistryOp(ctx, docker.RegistryOpPull, ptr(pullSandboxOK), "cr.example.com/proj/img:tag", now.Add(-50*time.Millisecond), nil)
	docker.RecordRegistryOp(ctx, docker.RegistryOpPull, nil, "alpine:3.20", now.Add(-40*time.Millisecond), nil)
	docker.RecordRegistryOp(ctx, docker.RegistryOpPull, ptr(pullSandboxOK), "cr.example.com/proj/img:tag", now.Add(-30*time.Millisecond), errors.New("simulated pull failure"))
	docker.RecordRegistryOp(ctx, docker.RegistryOpPush, ptr(pushSandboxOK), "localhost:5000/img:t", now.Add(-20*time.Millisecond), nil)
	docker.RecordRegistryOp(ctx, docker.RegistryOpPush, ptr(pushSandboxFailed), "registry.internal:5000/p/img", now.Add(-10*time.Millisecond), errors.New("boom"))

	// Shutdown flushes the PeriodicReader and blocks until exporters drain.
	require.NoError(t, shutdownOnce(), "shutdown should flush metrics without error")

	// Dump everything we received to the test log so `go test -v` shows the
	// exact wire payload — handy when iterating on instrumentation.
	srv.Dump(t)

	// ── assert ───────────────────────────────────────────────────────────

	t.Run("resource attributes carry service.name and namespace", func(t *testing.T) {
		attrs := srv.ResourceAttrs()
		require.Equal(t, "daytona-runner", attrs["service.name"], "service.name should be daytona-runner")
		require.Equal(t, "runner", attrs["service.namespace"], "service.namespace should be runner")
		require.Equal(t, "test", attrs["service.version"], "service.version should be the passed string")
		require.Equal(t, "test", attrs["deployment.environment.name"], "deployment.environment.name should be the passed string")

		// The resource is intentionally minimal. Keep this list locked down so
		// we don't accidentally re-enable the kitchen-sink detectors that
		// previously inflated every series with os/sdk/process labels.
		for _, k := range []string{
			"telemetry.sdk.name",
			"telemetry.sdk.language",
			"telemetry.sdk.version",
			"os.type",
			"os.description",
			"process.pid",
			"process.runtime.name",
			"process.runtime.version",
		} {
			require.NotContains(t, attrs, k, "resource should not carry %q", k)
		}
	})

	t.Run("registered metric instrument names are present", func(t *testing.T) {
		names := srv.MetricNames()
		require.Contains(t, names, "docker.registry.operation.count")
		require.Contains(t, names, "docker.registry.operation.duration")
		require.Contains(t, names, "docker.registry.operation.errors")
	})

	const (
		pullImageOK   = "cr.example.com/proj/img:tag"
		pullImageBare = "alpine:3.20"
		pushImageOK   = "localhost:5000/img:t"
		pushImageBad  = "registry.internal:5000/p/img"
	)

	t.Run("pull success with sandbox.id and image.ref", func(t *testing.T) {
		v, ok := srv.SumCounter("docker.registry.operation.count", map[string]string{
			"operation":     "pull",
			"status":        "success",
			"sandbox.id":    pullSandboxOK,
			"registry.host": "cr.example.com",
			"image.ref":     pullImageOK,
		})
		require.True(t, ok, "no matching counter data point received")
		require.EqualValues(t, 1, v, "counter should have observed exactly one success")

		hc, ok := srv.HistogramCount("docker.registry.operation.duration", map[string]string{
			"operation":     "pull",
			"sandbox.id":    pullSandboxOK,
			"registry.host": "cr.example.com",
			"image.ref":     pullImageOK,
		})
		require.True(t, ok, "no matching histogram data point received")
		require.GreaterOrEqual(t, hc, uint64(1), "histogram should have at least one observation")
	})

	t.Run("pull success without sandbox.id uses docker.io and omits attribute", func(t *testing.T) {
		v, ok := srv.SumCounter("docker.registry.operation.count", map[string]string{
			"operation":     "pull",
			"status":        "success",
			"registry.host": "docker.io",
			"image.ref":     pullImageBare,
		})
		require.True(t, ok, "no matching counter data point received")
		require.EqualValues(t, 1, v)

		require.True(t,
			srv.AnyDataPointMissing("docker.registry.operation.count", map[string]string{
				"operation":     "pull",
				"status":        "success",
				"registry.host": "docker.io",
				"image.ref":     pullImageBare,
			}, "sandbox.id"),
			"sandbox.id should be absent when caller passed nil",
		)
	})

	t.Run("pull failure increments both count and errors", func(t *testing.T) {
		v, ok := srv.SumCounter("docker.registry.operation.count", map[string]string{
			"operation":     "pull",
			"status":        "failure",
			"sandbox.id":    pullSandboxOK,
			"registry.host": "cr.example.com",
			"image.ref":     pullImageOK,
		})
		require.True(t, ok, "no matching failure counter data point received")
		require.EqualValues(t, 1, v)

		ev, ok := srv.SumCounter("docker.registry.operation.errors", map[string]string{
			"operation":     "pull",
			"status":        "failure",
			"sandbox.id":    pullSandboxOK,
			"registry.host": "cr.example.com",
			"image.ref":     pullImageOK,
		})
		require.True(t, ok, "no matching errors counter data point received")
		require.EqualValues(t, 1, ev)
	})

	t.Run("push success with localhost:5000 registry host", func(t *testing.T) {
		v, ok := srv.SumCounter("docker.registry.operation.count", map[string]string{
			"operation":     "push",
			"status":        "success",
			"sandbox.id":    pushSandboxOK,
			"registry.host": "localhost:5000",
			"image.ref":     pushImageOK,
		})
		require.True(t, ok, "no matching counter data point received")
		require.EqualValues(t, 1, v)
	})

	t.Run("push failure with hostname:port registry", func(t *testing.T) {
		v, ok := srv.SumCounter("docker.registry.operation.count", map[string]string{
			"operation":     "push",
			"status":        "failure",
			"sandbox.id":    pushSandboxFailed,
			"registry.host": "registry.internal:5000",
			"image.ref":     pushImageBad,
		})
		require.True(t, ok, "no matching counter data point received")
		require.EqualValues(t, 1, v)

		ev, ok := srv.SumCounter("docker.registry.operation.errors", map[string]string{
			"operation":     "push",
			"status":        "failure",
			"sandbox.id":    pushSandboxFailed,
			"registry.host": "registry.internal:5000",
			"image.ref":     pushImageBad,
		})
		require.True(t, ok, "no matching errors counter data point received")
		require.EqualValues(t, 1, ev)
	})

	t.Run("success-only operations did not increment the errors counter", func(t *testing.T) {
		_, ok := srv.SumCounter("docker.registry.operation.errors", map[string]string{
			"operation":     "push",
			"status":        "success",
			"sandbox.id":    pushSandboxOK,
			"registry.host": "localhost:5000",
			"image.ref":     pushImageOK,
		})
		require.False(t, ok, "errors counter must not record success operations")
	})
}

func ptr[T any](v T) *T {
	return &v
}
