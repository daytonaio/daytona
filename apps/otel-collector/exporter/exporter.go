// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package exporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/daytonaio/apiclient"
	"github.com/daytonaio/otel-collector/exporter/internal/config"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type IExporter[T any] interface {
	push(context.Context, T) error
	exportViaHTTP(context.Context, T, *apiclient.OtelConfig) error
	extractSandboxToken(context.Context) (string, error)
	getBody(T) ([]byte, error)
	shutdown(context.Context) error
}

type Exporter[T any] struct {
	config   *Config
	resolver *config.Resolver
	logger   *zap.Logger
	route    string

	httpClients map[string]*http.Client //lint:ignore U1000 Used in private methods consumed by the built collector
	mu          sync.RWMutex            //lint:ignore U1000 Used in private methods consumed by the built collector
	_getBody    func(T) ([]byte, error)
}

type exporterConfig struct {
	config   *Config
	logger   *zap.Logger
	resolver *config.Resolver
}

func newMetricExporter(cfg exporterConfig) IExporter[pmetric.Metrics] {
	return &Exporter[pmetric.Metrics]{
		config:   cfg.config,
		resolver: cfg.resolver,
		logger:   cfg.logger,
		route:    "v1/metrics",
		_getBody: func(md pmetric.Metrics) ([]byte, error) {
			req := pmetricotlp.NewExportRequestFromMetrics(md)
			return req.MarshalProto()
		},
	}
}

func newLogsExporter(cfg exporterConfig) IExporter[plog.Logs] {
	return &Exporter[plog.Logs]{
		config:   cfg.config,
		resolver: cfg.resolver,
		logger:   cfg.logger,
		route:    "v1/logs",
		_getBody: func(ld plog.Logs) ([]byte, error) {
			req := plogotlp.NewExportRequestFromLogs(ld)
			return req.MarshalProto()
		},
	}
}

func newTracesExporter(cfg exporterConfig) IExporter[ptrace.Traces] {
	return &Exporter[ptrace.Traces]{
		config:   cfg.config,
		resolver: cfg.resolver,
		logger:   cfg.logger,
		route:    "v1/traces",
		_getBody: func(td ptrace.Traces) ([]byte, error) {
			req := ptraceotlp.NewExportRequestFromTraces(td)
			return req.MarshalProto()
		},
	}
}

//lint:ignore U1000 Used by the built collector
func (e *Exporter[T]) push(ctx context.Context, data T) error {
	// Extract sandbox token from context metadata
	sandboxToken, err := e.extractSandboxToken(ctx)
	if err != nil {
		return consumererror.NewPermanent(fmt.Errorf("failed to extract sandbox token: %w", err))
	}

	// Get endpoint configuration
	endpointConfig, err := e.resolver.GetOrganizationOtelConfig(ctx, sandboxToken)
	if err != nil {
		return fmt.Errorf("failed to get endpoint config for sandbox %w", err)
	}

	if endpointConfig == nil {
		e.logger.Debug("No endpoint configuration found for sandbox token, dropping data")
		return nil
	}

	e.logger.Debug("Exporting data",
		zap.String("protocol", "http"),
		zap.String("endpoint", endpointConfig.Endpoint),
	)

	// Route to appropriate protocol handler
	return e.exportViaHTTP(ctx, data, endpointConfig)
}

//lint:ignore U1000 Used by the built collector
func (e *Exporter[T]) getBody(data T) ([]byte, error) {
	return e._getBody(data)
}

//lint:ignore U1000 Used by the built collector
func (e *Exporter[T]) exportViaHTTP(ctx context.Context, data T, cfg *apiclient.OtelConfig) error {
	httpClient := e.getOrCreateHTTPClient(cfg)

	// Create OTLP request and marshal to protobuf
	body, err := e.getBody(data)
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	// Create HTTP request
	endpoint := cfg.Endpoint
	if endpoint[len(endpoint)-1] != '/' {
		endpoint += "/"
	}
	endpoint += e.route

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	for k, v := range cfg.Headers {
		httpReq.Header.Set(k, v)
	}

	// Send request
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (e *Exporter[T]) getOrCreateHTTPClient(cfg *apiclient.OtelConfig) *http.Client {
	e.mu.RLock()
	if e.httpClients == nil {
		e.mu.RUnlock()
		e.mu.Lock()
		e.httpClients = make(map[string]*http.Client)
		e.mu.Unlock()
		e.mu.RLock()
	}

	client, exists := e.httpClients[cfg.Endpoint]
	e.mu.RUnlock()

	if exists {
		return client
	}

	// Create new HTTP client
	e.mu.Lock()
	defer e.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := e.httpClients[cfg.Endpoint]; exists {
		return client
	}

	client = &http.Client{
		Transport: http.DefaultTransport,
	}

	e.httpClients[cfg.Endpoint] = client

	return client
}

func (e *Exporter[T]) extractSandboxToken(ctx context.Context) (string, error) {
	// Try to get client info first (contains HTTP headers)
	clientInfo := client.FromContext(ctx)
	if token := clientInfo.Metadata.Get(e.config.SandboxAuthTokenHeader); len(token) > 0 {
		return token[0], nil
	}

	// Fallback: try gRPC metadata (if using gRPC protocol)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if tokens := md.Get(e.config.SandboxAuthTokenHeader); len(tokens) > 0 {
			return tokens[0], nil
		}
	}

	return "", fmt.Errorf("sandbox token header '%s' not found in metadata", e.config.SandboxAuthTokenHeader)
}

//lint:ignore U1000 Used by the built collector
func (e *Exporter[T]) shutdown(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.httpClients == nil {
		return nil
	}

	for _, client := range e.httpClients {
		// Close idle connections
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}

	e.httpClients = nil

	return nil
}
