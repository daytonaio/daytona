// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ProxyPort     int          `envconfig:"PROXY_PORT" validate:"required"`
	ProxyDomain   string       `envconfig:"PROXY_DOMAIN" validate:"required"`
	ProxyProtocol string       `envconfig:"PROXY_PROTOCOL" validate:"required"`
	ProxyApiKey   string       `envconfig:"PROXY_API_KEY" validate:"required"`
	TLSCertFile   string       `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile    string       `envconfig:"TLS_KEY_FILE"`
	EnableTLS     bool         `envconfig:"ENABLE_TLS"`
	DaytonaApiUrl string       `envconfig:"DAYTONA_API_URL" validate:"required"`
	Oidc          OidcConfig   `envconfig:"OIDC"`
	Redis         *RedisConfig `envconfig:"REDIS"`
}

type OidcConfig struct {
	ClientId     string `envconfig:"CLIENT_ID" validate:"required"`
	ClientSecret string `envconfig:"CLIENT_SECRET"`
	Domain       string `envconfig:"DOMAIN" validate:"required"`
	Audience     string `envconfig:"AUDIENCE" validate:"required"`
}

type RedisConfig struct {
	Host     *string `envconfig:"HOST"`
	Port     *int    `envconfig:"PORT"`
	Password *string `envconfig:"PASSWORD"`
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

	if config.Redis != nil {
		if config.Redis.Host == nil && config.Redis.Port == nil && config.Redis.Password == nil {
			config.Redis = nil
		}
	}

	return config, nil
}
