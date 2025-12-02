package exporter

import (
	"time"

	"go.opentelemetry.io/collector/config/configretry"
)

// Config defines the configuration for the custom exporter.
type Config struct {
	// SandboxIDHeader is the HTTP header name that contains the sandbox ID.
	// Default: "sandboxId"
	SandboxIDHeader string `mapstructure:"sandbox_id_header"`

	// CacheTTL is the duration to cache endpoint configurations.
	// Default: 5m
	CacheTTL time.Duration `mapstructure:"cache_ttl"`

	// DefaultTimeout is the timeout for OTLP export requests.
	// Default: 30s
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`

	// RetrySettings defines the retry behavior for failed exports.
	RetrySettings configretry.BackOffConfig `mapstructure:"retry_on_failure"`

	ApiUrl string `mapstructure:"api_url"`
	ApiKey      string `mapstructure:"api_key"`
}

// Validate checks if the configuration is valid.
func (cfg *Config) Validate() error {
	// Add validation logic here if needed
	return nil
}
