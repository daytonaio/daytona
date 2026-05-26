// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package exporter

import (
	"errors"
	"time"

	"github.com/daytonaio/common-go/pkg/cache"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines the configuration for the custom exporter.
type Config struct {
	// SandboxAuthTokenHeader is the HTTP header name that contains the sandbox auth token.
	// Default: "sandbox-auth-token"
	SandboxAuthTokenHeader string `mapstructure:"sandbox_auth_token_header"`

	// OrganizationIdHeader is the HTTP header name that contains the organization ID.
	// Used as a fallback when sandbox auth token is not present (e.g. org-level metrics).
	// Default: "organization-id"
	OrganizationIdHeader string `mapstructure:"organization_id_header"`

	// CacheTTL is the duration to cache endpoint configurations.
	// Default: 5m
	CacheTTL time.Duration `mapstructure:"cache_ttl"`

	// DefaultTimeout is the timeout for OTLP export requests.
	// Default: 30s
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`

	// RetrySettings defines the retry behavior for failed exports.
	RetrySettings configretry.BackOffConfig `mapstructure:"retry_on_failure"`

	// SendingQueue configures the queueing and batching behavior for sending requests to Daytona API.
	SendingQueue exporterhelper.QueueBatchConfig `mapstructure:"sending_queue"`

	// Daytona API configuration.
	ApiUrl string `mapstructure:"api_url"`
	ApiKey string `mapstructure:"api_key"`

	// Optional Redis config for caching endpoint configurations.
	Redis *cache.RedisConfig `mapstructure:"redis"`
}

func (cfg *Config) Validate() error {
	if cfg.Redis != nil {
		mode := ""
		if cfg.Redis.Mode != nil {
			mode = *cfg.Redis.Mode
		}
		if mode == "cluster" {
			if cfg.Redis.ClusterNodes == nil || *cfg.Redis.ClusterNodes == "" {
				return errors.New("redis cluster_nodes is required when redis mode is cluster")
			}
		} else if cfg.Redis.Host == nil || *cfg.Redis.Host == "" {
			cfg.Redis = nil
		}
	}
	return nil
}
