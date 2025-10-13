// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// CreateRootContext creates a context with a root span that can be used
// as the parent for all other operations
func CreateRootContext(ctx context.Context, tracerProvider *sdktrace.TracerProvider, rootSpanName string) (context.Context, trace.Span) {
	tracer := tracerProvider.Tracer("")
	return tracer.Start(ctx, rootSpanName)
}

// EnsureTraceContext ensures the context has trace information
func EnsureTraceContext(ctx context.Context, fallbackOperation string) context.Context {
	if trace.SpanContextFromContext(ctx).IsValid() {
		return ctx
	}

	// Create a new root span if no trace context exists
	tracer := otel.Tracer("your-service")
	newCtx, _ := tracer.Start(ctx, fallbackOperation)
	return newCtx
}
