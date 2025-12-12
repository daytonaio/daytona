/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package config

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

// Config holds the runner configuration
type Config struct {
	// API Configuration
	ApiUrl   string `envconfig:"DAYTONA_API_URL" validate:"required"`
	ApiToken string `envconfig:"DAYTONA_RUNNER_TOKEN" validate:"required"`

	Domain string `envconfig:"RUNNER_DOMAIN" validate:"required,hostname"`

	// Job Polling Configuration
	PollTimeout time.Duration `envconfig:"POLL_TIMEOUT" default:"30s"`
	PollLimit   int           `envconfig:"POLL_LIMIT" default:"10" validate:"min=1,max=100"`

	// Healthcheck Configuration
	HealthcheckInterval time.Duration `envconfig:"HEALTHCHECK_INTERVAL" default:"30s" validate:"min=10s"`
	HealthcheckTimeout  time.Duration `envconfig:"HEALTHCHECK_TIMEOUT" default:"10s"`

	// Telemetry Configuration
	OtelEnabled bool `envconfig:"OTEL_ENABLED" default:"false"`

	// Proxy Configuration
	ProxyPort        uint          `envconfig:"PROXY_PORT" default:"8080"`
	ProxyTLSEnabled  bool          `envconfig:"PROXY_TLS_ENABLED" default:"false"`
	ProxyTLSCertFile string        `envconfig:"PROXY_TLS_CERT_FILE"`
	ProxyTLSKeyFile  string        `envconfig:"PROXY_TLS_KEY_FILE"`
	ProxyCacheTTL    time.Duration `envconfig:"PROXY_CACHE_TTL" default:"10m"`
	ProxyTargetPort  int           `envconfig:"PROXY_TARGET_PORT" default:"2280"`
	ProxyNetwork     string        `envconfig:"PROXY_NETWORK" default:"bridge"`
}

var config *Config

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	var validate = validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
