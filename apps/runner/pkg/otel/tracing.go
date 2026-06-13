// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracing(ctx context.Context, res *resource.Resource, wrappers []func(sdktrace.SpanExporter) sdktrace.SpanExporter) (shutdown func(context.Context) error, err error) {
	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("otel: create trace exporter: %w", err)
	}

	for _, w := range wrappers {
		if w != nil {
			exporter = w(exporter)
		}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
