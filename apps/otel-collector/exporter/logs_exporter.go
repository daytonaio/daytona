package exporter

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sync"

	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/daytonaio/apiclient"
	"github.com/daytonaio/otel-collector/exporter/internal/cache"
	"github.com/daytonaio/otel-collector/exporter/internal/config"
)

type logsExporter struct {
	config   *Config
	resolver *config.Resolver
	logger   *zap.Logger

	// Connection pool for gRPC clients
	grpcClients map[string]plogotlp.GRPCClient
	httpClients map[string]*http.Client
	mu          sync.RWMutex
}

func (e *logsExporter) pushLogs(ctx context.Context, ld plog.Logs) error {
	// Extract sandbox ID from context metadata
	sandboxID, err := e.extractSandboxID(ctx)
	if err != nil {
		return consumererror.NewPermanent(fmt.Errorf("failed to extract sandbox ID: %w", err))
	}

	// Get endpoint configuration
	endpointConfig, err := e.resolver.GetOrganizationOtelConfig(ctx, sandboxID)
	if err != nil {
		return fmt.Errorf("failed to get endpoint config for sandbox %s: %w", sandboxID, err)
	}

	e.logger.Debug("Exporting logs",
		zap.String("sandbox_id", sandboxID),
		zap.String("customer_id", endpointConfig.CustomerID),
		zap.String("protocol", endpointConfig.Protocol),
		zap.String("endpoint", endpointConfig.Endpoint),
	)

	// Route to appropriate protocol handler
	return e.exportViaHTTP(ctx, ld, endpointConfig)
}

func (e *logsExporter) exportViaHTTP(ctx context.Context, ld plog.Logs, cfg *apiclient.OtelConfig) error {
	httpClient := e.getOrCreateHTTPClient(cfg)

	// Create OTLP request and marshal to protobuf
	req := plogotlp.NewExportRequestFromLogs(ld)
	body, err := req.MarshalProto()
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	// Create HTTP request
	endpoint := cfg.Endpoint
	if endpoint[len(endpoint)-1] != '/' {
		endpoint += "/"
	}
	endpoint += "v1/logs"

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

func (e *logsExporter) getOrCreateGRPCClient(cfg *cache.EndpointConfig) (plogotlp.GRPCClient, error) {
	e.mu.RLock()
	if e.grpcClients == nil {
		e.mu.RUnlock()
		e.mu.Lock()
		e.grpcClients = make(map[string]plogotlp.GRPCClient)
		e.mu.Unlock()
		e.mu.RLock()
	}

	client, exists := e.grpcClients[cfg.Endpoint]
	e.mu.RUnlock()

	if exists {
		return client, nil
	}

	// Create new gRPC client
	e.mu.Lock()
	defer e.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := e.grpcClients[cfg.Endpoint]; exists {
		return client, nil
	}

	// Set up gRPC connection options
	var opts []grpc.DialOption

	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	// Create connection
	conn, err := grpc.NewClient(cfg.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	client = plogotlp.NewGRPCClient(conn)
	e.grpcClients[cfg.Endpoint] = client

	return client, nil
}

func (e *logsExporter) getOrCreateHTTPClient(cfg *cache.EndpointConfig) *http.Client {
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

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
	}

	client = &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}

	e.httpClients[cfg.Endpoint] = client

	return client
}

func (e *logsExporter) extractSandboxID(ctx context.Context) (string, error) {
	// Extract from gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if values := md.Get(e.config.SandboxIDHeader); len(values) > 0 {
			return values[0], nil
		}
	}

	// TODO: Also check HTTP headers if this is an HTTP request
	// This might require additional context propagation from the receiver

	return "", fmt.Errorf("sandbox ID header '%s' not found in metadata", e.config.SandboxIDHeader)
}

func (e *logsExporter) addHeadersToContext(ctx context.Context, headers map[string]string) context.Context {
	if len(headers) == 0 {
		return ctx
	}

	md := metadata.New(headers)
	return metadata.NewOutgoingContext(ctx, md)
}

func (e *logsExporter) shutdown(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Close all gRPC connections
	e.grpcClients = nil
	e.httpClients = nil

	return nil
}
