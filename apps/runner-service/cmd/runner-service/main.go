/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner-service/internal/config"
	"github.com/daytonaio/runner-service/internal/daemon"
	"github.com/daytonaio/runner-service/internal/executor"
	"github.com/daytonaio/runner-service/internal/healthcheck"
	"github.com/daytonaio/runner-service/internal/logger"
	"github.com/daytonaio/runner-service/internal/metrics"
	"github.com/daytonaio/runner-service/internal/poller"
	"github.com/daytonaio/runner-service/internal/proxy"
	"github.com/daytonaio/runner-service/internal/telemetry"
)

func main() {
	// Setup structured logger
	log := logger.NewLogger()

	log.Info("Starting Daytona Runner",
		slog.String("version", "3"),
		slog.String("type", "job-based"))

	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Error("Failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("Configuration loaded",
		slog.String("api_url", cfg.ApiUrl),
		slog.Duration("poll_timeout", cfg.PollTimeout),
		slog.Int("poll_limit", cfg.PollLimit),
		slog.Duration("healthcheck_interval", cfg.HealthcheckInterval),
		slog.Bool("otel_enabled", cfg.OtelEnabled),
		slog.String("proxy_addr", cfg.ProxyAddr))

	// Initialize OpenTelemetry tracing if enabled
	var tp trace.TracerProvider
	if cfg.OtelEnabled {
		tracerProvider, err := telemetry.InitTracer()
		if err != nil {
			log.Error("Failed to initialize tracer", "error", err)
		} else {
			tp = tracerProvider
			defer telemetry.ShutdownTracer(tracerProvider)
		}
	}

	// Create API client with OpenTelemetry instrumentation
	// The otelhttp.NewTransport automatically propagates trace context via HTTP headers
	apiCfg := apiclient.NewConfiguration()
	apiCfg.HTTPClient = &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			}),
			otelhttp.WithSpanOptions(trace.WithSpanKind(trace.SpanKindClient)),
		),
	}
	apiCfg.Servers = apiclient.ServerConfigurations{
		{
			URL: cfg.ApiUrl,
		},
	}
	// Add API key as Bearer token (passport-http-bearer)
	apiCfg.AddDefaultHeader("Authorization", "Bearer "+cfg.ApiToken)

	log.Debug("API client configured",
		slog.String("api_url", cfg.ApiUrl),
		slog.String("auth_header", "Bearer "+cfg.ApiToken[:8]+"..."))

	apiClient := apiclient.NewAPIClient(apiCfg)

	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(), client.WithTraceProvider(tp))
	if err != nil {
		log.Error("Failed to create Docker client", slog.Any("error", err))
		os.Exit(1)
	}
	defer dockerClient.Close()

	log.Info("Docker client initialized")

	// Extract daemon binary on startup
	daemonPath, err := daemon.WriteStaticBinary("daemon-amd64")
	if err != nil {
		log.Error("Failed to write daemon binary", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("Daemon binary extracted", slog.String("path", daemonPath))

	// Create metrics collector
	metricsCollector := metrics.NewCollector(log)

	// Create executor with Docker client and daemon path
	jobExecutor := executor.NewExecutor(apiClient, dockerClient, metricsCollector, daemonPath, log)

	// Create services
	healthcheckService := healthcheck.NewService(cfg, apiClient, metricsCollector, log)
	pollerService := poller.NewService(cfg, apiClient, jobExecutor, log)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register signal handler early, before starting any services
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create proxy server
	proxyServer := proxy.NewProxyServer(proxy.ProxyServerConfig{
		Addr:         cfg.ProxyAddr,
		Log:          log,
		UseTLS:       cfg.ProxyTLSEnabled,
		CertFile:     cfg.ProxyTLSCertFile,
		KeyFile:      cfg.ProxyTLSKeyFile,
		CacheTTL:     cfg.ProxyCacheTTL,
		TargetPort:   cfg.ProxyTargetPort,
		Network:      cfg.ProxyNetwork,
		DockerClient: dockerClient,
	})

	// Start all services
	go func() {
		if err := proxyServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Error("Proxy server error", slog.Any("error", err))
		}
	}()

	go healthcheckService.Start(ctx)

	log.Info("Runner started successfully")
	go pollerService.Start(ctx)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Info("Received shutdown signal", slog.String("signal", sig.String()))
	log.Info("Initiating graceful shutdown")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown proxy server
	if err := proxyServer.Shutdown(shutdownCtx); err != nil {
		log.Error("Error shutting down proxy server", slog.Any("error", err))
	}

	// Cancel main context to stop healthcheck and poller
	cancel()

	log.Info("Runner shutdown complete")
}
