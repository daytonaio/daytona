// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/daytonaio/common-go/pkg/cache"
	"github.com/daytonaio/common-go/pkg/utils"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ProxyPort             int                `envconfig:"PROXY_PORT" validate:"required"`
	MetricsPort           int                `envconfig:"METRICS_PORT"`
	ProxyProtocol         string             `envconfig:"PROXY_PROTOCOL" validate:"required"`
	ProxyApiKey           string             `envconfig:"PROXY_API_KEY" validate:"required"`
	CookieDomain          *string            `envconfig:"COOKIE_DOMAIN"`
	TLSCertFile           string             `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile            string             `envconfig:"TLS_KEY_FILE"`
	EnableTLS             bool               `envconfig:"ENABLE_TLS"`
	DaytonaApiUrl         string             `envconfig:"DAYTONA_API_URL" validate:"required"`
	Oidc                  OidcConfig         `envconfig:"OIDC"`
	Redis                 *cache.RedisConfig `envconfig:"REDIS"`
	ToolboxOnlyMode       bool               `envconfig:"TOOLBOX_ONLY_MODE"`
	PreviewWarningEnabled bool               `envconfig:"PREVIEW_WARNING_ENABLED"`
	ShutdownTimeoutSec    int                `envconfig:"SHUTDOWN_TIMEOUT_SEC"`
	ApiClientTimeoutSec   int                `envconfig:"API_CLIENT_TIMEOUT_SEC"`
	ApiClient             *apiclient.APIClient
	// ApiHTTPTransport is the shared transport for HTTP clients that talk to the
	// Daytona API. Tighter IdleConnTimeout than http.DefaultTransport so we
	// don't reuse a connection the API server has already closed.
	ApiHTTPTransport http.RoundTripper
}

type OidcConfig struct {
	ClientId     string  `envconfig:"CLIENT_ID"`
	ClientSecret string  `envconfig:"CLIENT_SECRET"`
	Domain       string  `envconfig:"DOMAIN"`
	PublicDomain *string `envconfig:"PUBLIC_DOMAIN"`
	Audience     string  `envconfig:"AUDIENCE"`
}

var DEFAULT_PROXY_PORT int = 4000

const defaultApiClientTimeout = 60 * time.Second

// ApiClientTimeout returns the configured API client timeout, falling back to
// the default when unset so a zero value can never produce an already-expired
// context (which would silently drop API calls).
func (c *Config) ApiClientTimeout() time.Duration {
	if c.ApiClientTimeoutSec <= 0 {
		return defaultApiClientTimeout
	}
	return time.Duration(c.ApiClientTimeoutSec) * time.Second
}

var config *Config

func GetConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{}

	// Load .env files
	err := godotenv.Overload(".env", ".env.local", ".env.production")
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
		// Continue anyway, as environment variables might be set directly
	}

	err = envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	var validate = validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, err
	}

	if config.ProxyPort == 0 {
		config.ProxyPort = DEFAULT_PROXY_PORT
	}

	if config.ShutdownTimeoutSec == 0 {
		config.ShutdownTimeoutSec = 60 * 60 // default to 1 hour
	}

	if config.Redis != nil {
		if config.Redis.Host == nil || *config.Redis.Host == "" {
			config.Redis = nil
		}
	}

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: config.DaytonaApiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+config.ProxyApiKey)

	config.ApiClient = apiclient.NewAPIClient(clientConfig)

	// Clone http.DefaultTransport so we inherit its dial/TLS/H2 defaults, then
	// shorten IdleConnTimeout so we don't race the API server closing idle keep-alives.
	// Guard the type assertion so an unexpected http.DefaultTransport replacement
	// can't panic at startup; fall back to the current DefaultTransport as-is so we
	// don't silently lose its proxy/H2/dial defaults by using an empty transport.
	apiTransport := http.DefaultTransport
	if dt, ok := http.DefaultTransport.(*http.Transport); ok {
		cloned := dt.Clone()
		cloned.IdleConnTimeout = 30 * time.Second
		apiTransport = cloned
	} else {
		log.Println("Warning: http.DefaultTransport is not *http.Transport; using it as-is for the API client transport")
	}
	config.ApiHTTPTransport = apiTransport

	config.ApiClient.GetConfig().HTTPClient = &http.Client{
		Transport: config.ApiHTTPTransport,
		Timeout:   config.ApiClientTimeout(),
	}

	ctx := context.Background()

	// Retry fetching Daytona API config with exponential backoff
	err = utils.RetryWithExponentialBackoff(
		ctx,
		"get Daytona API config",
		10,
		time.Second,
		1*time.Minute,
		func() error {
			apiConfig, _, err := config.ApiClient.ConfigAPI.ConfigControllerGetConfig(ctx).Execute()
			if err != nil {
				return err
			}

			if config.Oidc.ClientId == "" {
				config.Oidc.ClientId = apiConfig.Oidc.ClientId
			}

			if config.Oidc.Domain == "" {
				config.Oidc.Domain = apiConfig.Oidc.Issuer

				if !strings.HasSuffix(config.Oidc.Domain, "/") {
					config.Oidc.Domain += "/"
				}
			}

			if config.Oidc.Audience == "" {
				config.Oidc.Audience = apiConfig.Oidc.Audience
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}
