// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ApiUrl          string `json:"apiUrl" envconfig:"DAYTONA_API_URL" validate:"required"`
	Port            int    `json:"port" envconfig:"PORT" validate:"required" default:"8080"`
	TLSCertFilePath string `json:"tlsCertFilePath" envconfig:"TLS_CERT_FILE_PATH" validate:"required"`
	TLSKeyFilePath  string `json:"tlsKeyFilePath" envconfig:"TLS_KEY_FILE_PATH" validate:"required"`
	Auth0Domain     string `json:"auth0Domain" envconfig:"DAYTONA_AUTH0_DOMAIN" validate:"required"`
	Auth0ClientId   string `json:"auth0ClientId" envconfig:"DAYTONA_AUTH0_CLIENT_ID" validate:"required"`
	Auth0Audience   string `json:"auth0Audience" envconfig:"DAYTONA_AUTH0_AUDIENCE" validate:"required"`
}

var config *Config

func GetConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
