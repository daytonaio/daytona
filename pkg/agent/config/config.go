// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	ProjectDir string `envconfig:"DAYTONA_WS_DIR" validate:"required"`
	Server     struct {
		Url    string `envconfig:"DAYTONA_SERVER_URL" validate:"required"`
		ApiKey string `envconfig:"DAYTONA_SERVER_API_KEY" validate:"required"`
	}
}

var config *Config

func GetConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}

	var validate = validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
