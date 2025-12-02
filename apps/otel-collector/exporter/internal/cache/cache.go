package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching customer endpoint configurations.
// This allows for easy swapping between in-memory and Redis implementations.
type Cache interface {
	// Get retrieves a cached configuration for the given sandbox ID.
	// Returns nil if not found or expired.
	Get(ctx context.Context, sandboxID string) (*EndpointConfig, error)

	// Set stores a configuration for the given sandbox ID with TTL.
	Set(ctx context.Context, sandboxID string, config *EndpointConfig, ttl time.Duration) error

	// Delete removes a cached configuration for the given sandbox ID.
	Delete(ctx context.Context, sandboxID string) error

	// Clear removes all cached configurations.
	Clear(ctx context.Context) error
}

// EndpointConfig represents the customer's OTLP endpoint configuration.
type EndpointConfig struct {
	CustomerID string
	SandboxID  string

	// Protocol can be "http" or "grpc"
	Protocol string

	// Endpoint is the OTLP endpoint URL
	Endpoint string

	// Headers contains authentication and custom headers
	Headers map[string]string

	// TLS configuration
	Insecure           bool
	InsecureSkipVerify bool

	// Timeout for requests
	Timeout time.Duration

	// Cached timestamp
	CachedAt time.Time
}
