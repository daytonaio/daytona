// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"log"

	"github.com/daytonaio/common-go/pkg/cache"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port          int                `envconfig:"PORT" validate:"required"`
	ApiKey        string             `envconfig:"API_KEY" validate:"required"`
	TLSCertFile   string             `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile    string             `envconfig:"TLS_KEY_FILE"`
	EnableTLS     bool               `envconfig:"ENABLE_TLS"`
	DaytonaApiUrl string             `envconfig:"DAYTONA_API_URL" validate:"required"`
	Redis         *cache.RedisConfig `envconfig:"REDIS"`
}

var DEFAULT_PORT int = 8000

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

	if config.Port == 0 {
		config.Port = DEFAULT_PORT
	}

	if config.Redis != nil {
		if config.Redis.Host == nil && config.Redis.Port == nil && config.Redis.Password == nil {
			config.Redis = nil
		}
	}

	return config, nil
}
