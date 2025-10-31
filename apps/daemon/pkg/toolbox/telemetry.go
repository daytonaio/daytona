// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"context"
	"fmt"

	"github.com/daytonaio/daemon/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *server) initTelemetry(ctx context.Context, serviceName string) error {
	if s.telemetry.Logger != nil {
		if err := s.telemetry.Logger.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown existing telemetry logger: %w", err)
		}
	}

	if s.telemetry.MeterProvider != nil {
		if err := s.telemetry.MeterProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown existing telemetry meter provider: %w", err)
		}
	}

	if s.telemetry.TracerProvider != nil {
		if err := s.telemetry.TracerProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown existing telemetry tracer provider: %w", err)
		}
	}

	config := telemetry.Config{
		ServiceName: serviceName,
		Endpoint:    s.otelEndpoint,
		Headers: map[string]string{
			"sandbox-auth-token": s.authToken,
		},
	}

	// Use a background context
	telemetryContext := context.Background()

	// Initialize OpenTelemetry logging
	lp, err := telemetry.InitLogger(telemetryContext, config)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize OpenTelemetry metrics
	mp, err := telemetry.InitMetrics(ctx, config)
	if err != nil {
		defer lp.Shutdown(telemetryContext)
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize OpenTelemetry tracing
	tp, err := telemetry.InitTracer(ctx, config)
	if err != nil {
		defer lp.Shutdown(telemetryContext)
		defer mp.Shutdown(telemetryContext)
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}

	s.telemetry.TracerProvider = tp
	s.telemetry.MeterProvider = mp
	s.telemetry.Logger = lp

	log.Info("Telemetry initialized successfully")
	return nil
}
