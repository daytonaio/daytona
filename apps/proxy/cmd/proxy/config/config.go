// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/daytonaio/apiclient"
	"github.com/daytonaio/common-go/pkg/cache"
	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ProxyPort             int                `envconfig:"PROXY_PORT" validate:"required"`
	ProxyDomain           string             `envconfig:"PROXY_DOMAIN" validate:"required"`
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
	ApiClient             *apiclient.APIClient
}

type OidcConfig struct {
	ClientId     string  `envconfig:"CLIENT_ID"`
	ClientSecret string  `envconfig:"CLIENT_SECRET"`
	Domain       string  `envconfig:"DOMAIN"`
	PublicDomain *string `envconfig:"PUBLIC_DOMAIN"`
	Audience     string  `envconfig:"AUDIENCE"`
}

var DEFAULT_PROXY_PORT int = 4000

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
		if config.Redis.Host == nil && config.Redis.Port == nil && config.Redis.Password == nil {
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

	config.ApiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	ctx := context.Background()

	// Retry fetching Daytona API config with exponential backoff
	err = utils.RetryWithExponentialBackoff(
		ctx,
		"get Daytona API config",
		utils.DEFAULT_MAX_RETRIES,
		time.Second,
		10*time.Second,
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
