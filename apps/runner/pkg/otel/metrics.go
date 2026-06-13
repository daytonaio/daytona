// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func initMetrics(ctx context.Context, res *resource.Resource) (func(context.Context) error, error) {
	reader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("otel: create metric reader: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(15 * time.Second)); err != nil {
		return mp.Shutdown, fmt.Errorf("otel: start runtime metrics: %w", err)
	}
	if err := host.Start(host.WithMeterProvider(mp)); err != nil {
		return mp.Shutdown, fmt.Errorf("otel: start host metrics: %w", err)
	}

	return mp.Shutdown, nil
}
