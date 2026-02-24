// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"context"
	"fmt"

	"github.com/daytonaio/common-go/pkg/telemetry"
	"github.com/daytonaio/daemon/internal"
)

func (s *server) initTelemetry(ctx context.Context, serviceName string) error {
	if s.otelEndpoint == nil {
		s.logger.InfoContext(ctx, "Otel endpoint not provided, skipping telemetry initialization")
		return nil
	}

	if s.telemetry.LoggerProvider != nil {
		if err := s.telemetry.LoggerProvider.Shutdown(ctx); err != nil {
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
		ServiceName:    serviceName,
		ServiceVersion: internal.Version,
		Endpoint:       *s.otelEndpoint,
		Headers: map[string]string{
			"sandbox-auth-token": s.authToken,
		},
	}

	// Use a background context
	telemetryContext := context.Background()

	// Initialize OpenTelemetry logging
	newLogger, lp, err := telemetry.InitLogger(telemetryContext, s.logger, config)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	s.logger = newLogger

	// Initialize OpenTelemetry metrics
	mp, err := telemetry.InitMetrics(ctx, config, "daytona.sandbox")
	if err != nil {
		if shutDownErr := lp.Shutdown(telemetryContext); shutDownErr != nil {
			s.logger.ErrorContext(ctx, "Failed to shutdown logger after metrics initialization failure", "shutdownErr", shutDownErr)
		}
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize OpenTelemetry tracing
	tp, err := telemetry.InitTracer(ctx, config)
	if err != nil {
		if shutDownErr := lp.Shutdown(telemetryContext); shutDownErr != nil {
			s.logger.ErrorContext(ctx, "Failed to shutdown logger after tracer initialization failure", "shutdownErr", shutDownErr)
		}
		if shutDownErr := mp.Shutdown(telemetryContext); shutDownErr != nil {
			s.logger.ErrorContext(ctx, "Failed to shutdown meter provider after tracer initialization failure", "shutdownErr", shutDownErr)
		}
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}

	s.telemetry.TracerProvider = tp
	s.telemetry.MeterProvider = mp
	s.telemetry.LoggerProvider = lp

	s.logger.InfoContext(ctx, "Telemetry initialized successfully")
	return nil
}
