// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/common-go/pkg/telemetry"
	"github.com/daytonaio/daemon/internal"
)

func (s *server) initTelemetry(ctx context.Context, serviceName, entrypointLogFilePath string, organizationId, regionId, snapshot *string) error {
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

	extraLabels := make(map[string]string)
	if organizationId != nil && *organizationId != "" {
		extraLabels["daytona_organization_id"] = *organizationId
	}

	if regionId != nil && *regionId != "" {
		extraLabels["daytona_region_id"] = *regionId
	}

	if snapshot != nil && *snapshot != "" {
		extraLabels["daytona_snapshot"] = *snapshot
	}

	if envLabels := os.Getenv("DAYTONA_SANDBOX_OTEL_EXTRA_LABELS"); envLabels != "" {
		for pair := range strings.SplitSeq(envLabels, ",") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				s.logger.WarnContext(ctx, "Skipping malformed extra label", "label", pair)
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				extraLabels[key] = value
			}
		}
	}

	if len(extraLabels) > 0 {
		config.ExtraLabels = extraLabels
	}

	// Use a background context
	telemetryContext := context.Background()

	// Initialize OpenTelemetry logging
	newLogger, lp, err := telemetry.InitLogger(telemetryContext, s.logger, config)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	s.logger = newLogger

	if s.entrypointLogCancel != nil {
		s.entrypointLogCancel()
	}

	entrypointCtx, entrypointCancel := context.WithCancel(s.ctx)
	s.entrypointLogCancel = entrypointCancel

	go func() {
		if entrypointLogFilePath == "" {
			return
		}

		entrypointLogFile, err := os.Open(entrypointLogFilePath)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to open entrypoint log file", "error", err, "daytona-entrypoint", true)
			return
		}
		defer entrypointLogFile.Close()

		errChan := make(chan error, 1)
		stdoutChan := make(chan []byte)
		stderrChan := make(chan []byte)
		go log.ReadMultiplexedLog(entrypointCtx, entrypointLogFile, true, stdoutChan, stderrChan, errChan)
		for {
			select {
			case <-entrypointCtx.Done():
				return
			case line := <-stdoutChan:
				s.logger.InfoContext(telemetryContext, string(line), "daytona-entrypoint", true)
			case line := <-stderrChan:
				s.logger.ErrorContext(telemetryContext, string(line), "daytona-entrypoint", true)
			case err := <-errChan:
				if err != nil {
					s.logger.ErrorContext(telemetryContext, "Error reading entrypoint log file", "error", err, "daytona-entrypoint", true)
				}
				return
			}
		}
	}()

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
